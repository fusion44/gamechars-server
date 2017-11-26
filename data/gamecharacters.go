package data

import (
	"context"
	"fmt"

	"github.com/fusion44/gamechars-server/utils"
	graphql "github.com/neelance/graphql-go"
	"github.com/rs/xid"
)

// gameCharacter corresponds to the GQL type GameCharacter.
type gameCharacter struct {
	ID          graphql.ID
	Name        string
	DebutGame   string
	ReleaseYear int32
	Img         string
	Desc        string
	Wiki        string
	Public      bool
	Owner       string
}

type result struct {
	Op    string
	Count int32
}

// gameCharacters some hardcoded data to work with
var gameCharacters = []*gameCharacter{
	{
		ID:          "1000",
		Name:        "Gordon Freeman",
		DebutGame:   "Half-Life",
		ReleaseYear: 1998,
		Img:         "gordon_freeman.jpg",
		Desc:        "Dr. Gordon Freeman is a fictional character and the main protagonist of the Half-Life video game series, created by Gabe Newell[2] and designed by Newell and Marc Laidlaw[3] of Valve Corporation. His first appearance is in Half-Life. Gordon Freeman is an American man from Seattle, who graduated from MIT with a PhD in Theoretical Physics. He was an employee at Black Mesa Research Facility. Controlled by the player, Gordon is often tasked with using a wide range of weapons and tools to fight alien creatures such as headcrabs, as well as Combine machines and soldiers.",
		Wiki:        "https://en.wikipedia.org/wiki/Gordon_Freeman",
		Public:      true,
		Owner:       "fusion44",
	},
	{
		ID:          "1001",
		Name:        "GLaDOS",
		DebutGame:   "Portal",
		ReleaseYear: 2007,
		Img:         "glados.png",
		Desc:        "GLaDOS (Genetic Lifeform and Disk Operating System)[1] is a fictional artificially intelligent computer system from the video game series Portal.",
		Wiki:        "https://en.wikipedia.org/wiki/GLaDOS",
		Public:      true,
		Owner:       "fusion44",
	},
	{
		ID:          "1002",
		Name:        "Shodan",
		DebutGame:   "System Shock",
		ReleaseYear: 1994,
		Img:         "shodan.jpg",
		Desc:        "SHODAN (Sentient Hyper-Optimized Data Access Network) is a fictional artificial intelligence and the main antagonist of the cyberpunk-horror themed action role-playing video games System Shock and System Shock 2.",
		Wiki:        "https://en.wikipedia.org/wiki/SHODAN",
		Public:      true,
		Owner:       "fusion44",
	},
	{
		ID:          "1003",
		Name:        "The Nameless One",
		DebutGame:   "Planescape: Torment",
		ReleaseYear: 1999,
		Img:         "nameless_one.jpg",
		Desc:        "\"The Nameless One\" is a fictional character from the Black Isle Studios role-playing video game Planescape: Torment, and is the main protagonist of the story. The character was voiced by Michael T. Weiss and created by game designer Chris Avellone. The Nameless One is a heavily scarred immortal, who, when killed, may suffer severe memory loss.",
		Wiki:        "https://en.wikipedia.org/wiki/The_Nameless_One",
		Public:      true,
		Owner:       "fusion44",
	}, {
		ID:          "1004",
		Name:        "Lara Croft",
		DebutGame:   "Tomb Raider",
		ReleaseYear: 1996,
		Img:         "lara.jpg",
		Desc:        "Lara Croft is a fictional character and the main protagonist of the Square Enix (previously Eidos Interactive) video game franchise Tomb Raider. She is presented as a highly intelligent, athletic, and beautiful English archaeologist-adventurer who ventures into ancient, hazardous tombs and ruins around the world. Created by a team at UK developer Core Design that included Toby Gard, the character first appeared in the 1996 video game Tomb Raider. She has also appeared in video game sequels, printed adaptations, a series of animated short films, feature films (portrayed by Angelina Jolie, later by Alicia Vikander), and merchandise related to the series. Official promotion of the character includes a brand of apparel and accessories, action figures, and model portrayals. Croft has also been licensed for third-party promotion, including television and print advertisements, music-related appearances, and as a spokesmodel. As of June 2016, Lara Croft has been featured on over 1,100 magazine covers surpassing any supermodel.",
		Wiki:        "https://en.wikipedia.org/wiki/Lara_Croft",
		Public:      true,
		Owner:       "fusion44",
	}, {
		ID:          "1005",
		Name:        "Cate Archer",
		DebutGame:   "No One Lives Forever",
		ReleaseYear: 2000,
		Img:         "cate_archer.jpg",
		Desc:        "Catherine Ann \"Cate\" Archer, codenamed The Fox, is a player character and the protagonist in the No One Lives Forever video game series by Monolith Productions. Cate, a covert operative for British-based counter-terrorism organization UNITY, is the main character in The Operative: No One Lives Forever (2000) and No One Lives Forever 2: A Spy In H.A.R.M.'s Way (2002), and is also featured in Contract J.A.C.K., an official prequel to the second game.",
		Wiki:        "https://en.wikipedia.org/wiki/Cate_Archer",
		Public:      true,
		Owner:       "fusion44",
	},
}

var gameCharacterData = make(map[graphql.ID]*gameCharacter)

func init() {
	for _, h := range gameCharacters {
		gameCharacterData[h.ID] = h
	}
}

// Resolver type holds all the specialized resolvers that implement GQL queries and mutations
type Resolver struct{}

type resultResolver struct {
	result *result
}

func (gcr *resultResolver) Op() string {
	return gcr.result.Op
}

func (gcr *resultResolver) Count() int32 {
	return gcr.result.Count
}

// GameCharacter gets one character by its ID
func (r *Resolver) GameCharacter(ctx context.Context, args struct {
	ID graphql.ID
}) *gameCharacterResolver {
	auth, err := utils.GetContextAuthData(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}

	if gc := gameCharacterData[args.ID]; gc != nil {
		// only return the data if the character data is public
		// or the currently logged in user is marked as owner
		if gc.Public || gc.Owner == auth.UserName {
			return &gameCharacterResolver{gc}
		}
	}
	return nil
}

// GameCharacters gets all characters in the database
func (r *Resolver) GameCharacters(ctx context.Context) *[]*gameCharacterResolver {
	auth, err := utils.GetContextAuthData(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}

	var gameChars []*gameCharacterResolver
	for _, gc := range gameCharacterData {
		if gc.Public || gc.Owner == auth.UserName {
			gameChars = append(gameChars, &gameCharacterResolver{gc})
		}
	}

	return &gameChars
}

// gameCharacterResolver resolves individual fields of a game character
type gameCharacterResolver struct {
	gameCharacter *gameCharacter
}

func (gcr *gameCharacterResolver) ID() graphql.ID {
	return gcr.gameCharacter.ID
}

func (gcr *gameCharacterResolver) Name() string {
	return gcr.gameCharacter.Name
}

func (gcr *gameCharacterResolver) DebutGame() string {
	return gcr.gameCharacter.DebutGame
}

func (gcr *gameCharacterResolver) ReleaseYear() int32 {
	return gcr.gameCharacter.ReleaseYear
}

func (gcr *gameCharacterResolver) Img() string {
	return gcr.gameCharacter.Img
}

func (gcr *gameCharacterResolver) Desc() string {
	return gcr.gameCharacter.Desc
}

func (gcr *gameCharacterResolver) Wiki() string {
	return gcr.gameCharacter.Wiki
}

func (gcr *gameCharacterResolver) Public() bool {
	return gcr.gameCharacter.Public
}

func (gcr *gameCharacterResolver) Owner() string {
	return gcr.gameCharacter.Owner
}

type userResolver struct {
	user *user
}

func (u *userResolver) ID() graphql.ID {
	return u.user.ID
}

func (u *userResolver) UserName() string {
	return u.user.UserName
}

func (u *userResolver) Email() string {
	return u.user.Email
}

func (u *userResolver) Token() string {
	return u.user.Token
}

// A user that is signed in
type user struct {
	ID       graphql.ID
	UserName string
	Email    string
	Token    string
}

// UserDbModel is exported because the auth API requires this model
type UserDbModel struct {
	ID       graphql.ID
	UserName []byte
	Email    []byte
	Password []byte
}

// CHARACTERS
type gameCharacterInput struct {
	Name        string
	DebutGame   string
	ReleaseYear int32
	Img         string
	Desc        string
	Wiki        string
	Public      bool
	Owner       string
}

// AddCharacter Adds a new character to the database
func (r *Resolver) AddCharacter(args *struct {
	Char *gameCharacterInput // Argument name must be the same as in GQL
}) *gameCharacterResolver {
	gc := &gameCharacter{
		ID:          graphql.ID(xid.New().String()),
		Name:        args.Char.Name,
		DebutGame:   args.Char.DebutGame,
		ReleaseYear: args.Char.ReleaseYear,
		Img:         args.Char.Img,
		Desc:        args.Char.Desc,
		Wiki:        args.Char.Wiki,
		Public:      args.Char.Public,
		Owner:       args.Char.Owner,
	}
	gameCharacterData[gc.ID] = gc
	return &gameCharacterResolver{gc}
}

func (r *Resolver) RemoveCharacter(ctx context.Context, args *struct {
	ID graphql.ID
}) *resultResolver {
	res := result{
		Op:    "delete",
		Count: 0,
	}

	// Get authentication data
	auth, err := utils.GetContextAuthData(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return &resultResolver{&res}
	}

	// Check if there is data to delete
	before := len(gameCharacterData)
	if before < 1 {
		fmt.Println("No characters to delete")
		return &resultResolver{&res}
	}

	if gc := gameCharacterData[args.ID]; gc != nil {
		// Check the first
		if gc.Owner == auth.UserName {
			delete(gameCharacterData, args.ID)
			res.Count = int32(before - len(gameCharacterData))
		}
	}

	return &resultResolver{&res}
}
