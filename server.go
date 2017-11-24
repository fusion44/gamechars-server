package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	gorillaContext "github.com/gorilla/context"

	"github.com/gorilla/sessions"

	"github.com/coreos/bbolt"
	"github.com/fusion44/gamechars-server/data"
	"github.com/fusion44/gamechars-server/utils"
	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/nicksrandall/batched-graphql-handler"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

var schema *graphql.Schema
var store = sessions.NewCookieStore([]byte("something-very-secret"))
var validate *validator.Validate

const cookieName = "gamechars-session"
const userOpSuccessMsg = "OK"

func init() {
	gameCharacterSchema, err := ioutil.ReadFile("./data/gamecharacters.gql")
	check(err)

	schema = graphql.MustParseSchema(string(gameCharacterSchema), &data.Resolver{})
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func setSessionOnClient(w http.ResponseWriter, r *http.Request, userName string) {
	session, err := store.Get(r, cookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// https://gowebexamples.com/sessions/
	// the auth handler will read this value
	// logout will set this to false
	session.Values["authenticated"] = true
	session.Values["userName"] = userName

	// This will set the cookie in the client browser
	session.Save(r, w)

	// Return the username that was just signed in
	w.Header().Add("Content-Type", "application/json")
	msg, err := json.Marshal(userOpSuccessReturn{
		UserName: userName,
	})

	if err == nil {
		w.Write(msg)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error processing the request"))
	}
}

type userOpSuccessReturn struct {
	UserName string `json:"userName"`
}

type userInput struct {
	UserName string `validate:"required,min=2,max=16"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=4"`
}

type userInputValidationError struct {
	Type string
	Err  string `json:"error"`
}

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST ist allowed", 500)
		return
	}

	// Open BoltDB
	db, err := bolt.Open("gamechars.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Close the DB when this function returns
	defer db.Close()

	decoder := json.NewDecoder(r.Body)
	var uinput userInput
	check(decoder.Decode(&uinput))
	err = validate.Struct(uinput)
	defer r.Body.Close()

	if err != nil {
		var errors []userInputValidationError
		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "UserName" {
				errorString := fmt.Sprintf("Username %s is to short. Minimum length is two.", err.Value())
				errors = append(errors, userInputValidationError{Type: "userName", Err: errorString})
			}
			if err.Field() == "Email" {
				errorString := fmt.Sprintf("%s is not a valid email address.", err.Value())
				errors = append(errors, userInputValidationError{Type: "email", Err: errorString})
			}
			if err.Field() == "Password" {
				errorString := fmt.Sprintf("Password is to short. Minimum length is four.")
				errors = append(errors, userInputValidationError{Type: "password", Err: errorString})
			}
		}

		enc, err := json.Marshal(errors)
		check(err)
		http.Error(w, string(enc), http.StatusBadRequest)
		return
	}

	userFound := utils.CheckUserNameExists(uinput.UserName, db)

	if userFound {
		http.Error(w, "User name is taken", 303)
		return
	}

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Users"))

		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(uinput.Password), bcrypt.DefaultCost)

		if err != nil {
			log.Fatal("Unable to process password")
			return errors.Errorf("Unable to process password")
		}

		udb := data.UserDbModel{
			ID:       graphql.ID(xid.New().String()),
			UserName: []byte(uinput.UserName),
			Email:    []byte(uinput.Email),
			Password: hashedPassword,
		}

		udbJSON, err := json.Marshal(udb)
		if err != nil {
			log.Fatal("Unable marshal user to JSON")
			return errors.Errorf("Unable marshal user to JSON")
		}
		b.Put(udb.UserName, udbJSON)

		setSessionOnClient(w, r, uinput.UserName)

		fmt.Printf("User %s created.\n", uinput.UserName)

		return nil
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST ist allowed", 500)
		return
	}

	// Open BoltDB
	db, err := bolt.Open("gamechars.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Close the DB when this function returns
	defer db.Close()

	decoder := json.NewDecoder(r.Body)
	var uinput userInput
	check(decoder.Decode(&uinput))
	err = validate.Struct(uinput)
	defer r.Body.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Users"))
		entry := b.Get([]byte(uinput.UserName))
		if entry == nil {
			http.Error(w, "Username or password is wrong", 401)
			return nil
		}
		var dbUser data.UserDbModel

		json.Unmarshal(entry, &dbUser)
		err = bcrypt.CompareHashAndPassword(dbUser.Password, []byte(uinput.Password))
		if err == nil {
			// Username found and password is OK
			setSessionOnClient(w, r, uinput.UserName)
		} else {
			// Password is wrong, send error. Username is checked above.
			http.Error(w, "Username or password is wrong", 401)
			return nil
		}

		fmt.Printf("User %s logged in.\n", uinput.UserName)
		return nil
	})

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST ist allowed", 500)
		return
	}

	// Open BoltDB
	db, err := bolt.Open("gamechars.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Close the DB when this function returns
	defer db.Close()

	session, err := store.Get(r, cookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userName := session.Values["userName"].(string)

	// https://gowebexamples.com/sessions/
	// the auth handler will read this value
	// logout will set this to false
	session.Values["authenticated"] = false
	session.Values["userName"] = ""

	// This will set the cookie in the client browser
	err = session.Save(r, w)

	if err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{'status':'OK'}"))
		fmt.Printf("User %s logged out.\n", userName)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error processing the request"))
	}
}

func authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, cookieName)
		ctx := r.Context()
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		} else {
			auth := false
			userName := ""
			// Check if user is authenticated
			if session.Values["authenticated"] != nil {
				auth = session.Values["authenticated"].(bool)
				userName = session.Values["userName"].(string)
			}

			next.ServeHTTP(w, r.WithContext(utils.PutContextAuthData(ctx, auth, userName)))
		}
	})
}

func main() {
	db, err := bolt.Open("gamechars.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Users"))
		if err != nil {
			return fmt.Errorf("create users bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("GameCharacters"))
		if err != nil {
			return fmt.Errorf("create game characters bucket: %s", err)
		}

		return nil
	})

	db.Close()
	validate = validator.New()

	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 30,
	}

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
	})

	http.Handle("/auth/signup", gorillaContext.ClearHandler(c.Handler(http.HandlerFunc(signUpHandler))))

	http.Handle("/auth/login", gorillaContext.ClearHandler(c.Handler(http.HandlerFunc(loginHandler))))

	http.Handle("/auth/logout", gorillaContext.ClearHandler(c.Handler(http.HandlerFunc(logoutHandler))))

	http.Handle("/graphql", gorillaContext.ClearHandler(
		c.Handler(authHandler(&apollo.Handler{Schema: schema}))))
	http.Handle("/graphiql", &relay.Handler{Schema: schema})

	fmt.Println("Running Server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var page = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.css" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.5/graphiql.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/graphiql", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
