package game
import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var funcMap = map[string]interface{}{"joinList": createJoinList}
var joinTemplate = template.Must(template.New("join").Funcs(funcMap).ParseFiles("../client/join.html"))
var hostTemplate = template.Must(template.New("host").Funcs(funcMap).ParseFiles("../client/host.html"))

// TODO: add last time pinged to struct for player cleanup
type player struct {
	sessionID int
	id        int
	name      string
}

var currentPlayers = []player{player{0, 0, "player1"}, player{1, 1, "player2"}}
var joinList = []*player{&currentPlayers[0], &currentPlayers[1]}
var playerMux sync.Mutex

func createJoinList() template.HTML {

	selections := ""
	playerMux.Lock()
	defer playerMux.Unlock()

	for i := 0; i < len(joinList); i++ {
		selections += joinList[i].name + "  <button onclick=\"window.location='/join?id=" +
			fmt.Sprint(joinList[i].id) + "'\">Join</button><br>"
	}

	return template.HTML(selections)
}

func Join(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 0, 0)
	idIndex := getJoinIndex(int(id))
	if err != nil || idIndex == -1 {
		w.Header().Set("Content-type", "text/html; charset=utf-8")
		type data struct{ EnterName bool }
		err := joinTemplate.ExecuteTemplate(w, "join.html", data{false})
		return err
	}

	username := r.FormValue("username")
	// TODO: sanitize names for html
	if username == "" {
		w.Header().Set("Content-type", "text/html; charset=utf-8")
		type data struct{ EnterName bool }
		err := joinTemplate.ExecuteTemplate(w, "join.html", data{true})
		return err
	}

	sessionID, newID := newIDPair()

	playerMux.Lock()
	currentPlayers = append(currentPlayers, player{sessionID, newID, username})
	currentGames = append(currentGames, [2]*player{&currentPlayers[idIndex], &currentPlayers[len(currentPlayers)-1]})
	playerMux.Unlock()

	removePlayerFromJoin(int(id))
	http.Redirect(w, r, "/play", 301)
	return nil
}

func CheckForJoin(w http.ResponseWriter, r *http.Request) error {
	cookie, _ := r.Cookie("sessionID")
	sessionID, _ := strconv.ParseInt(cookie.Value, 0, 0)
	gameIndex, _ := getGameIndex(-1, int(sessionID))
	fmt.Fprint(w, gameIndex != -1)
	return nil
}

func Host(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("sessionID")

	var foundCookie bool
	var validSessionID bool
	if err != nil {
		foundCookie = false
	} else {
		foundCookie = true
		sessionID, err := strconv.ParseInt(cookie.Value, 0, 0)
		if err != nil {
			foundCookie = false
		}
		index := getSessionIDIndex(int(sessionID))
		validSessionID = index != -1
	}

	type data struct{ New bool }
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	err = hostTemplate.ExecuteTemplate(w, "host.html", data{!(foundCookie && validSessionID)})
	return err
}

func New(w http.ResponseWriter, r *http.Request) error {
	username := r.FormValue("username")
	// TODO: sanitize names for html
	sessionID, id := newIDPair()

	playerMux.Lock()
	defer playerMux.Unlock()

	currentPlayers = append(currentPlayers, player{sessionID, id, username})
	joinList = append(joinList, &currentPlayers[len(currentPlayers)-1])

	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{Name: "sessionID", Value: fmt.Sprint(sessionID), Expires: expiration}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/host", 301)
	return nil
}

func removePlayerFromJoin(id int) {
	index := getJoinIndex(id)
	length := len(joinList)

	playerMux.Lock()
	defer playerMux.Unlock()

	joinList[index] = joinList[length-1]
	joinList = joinList[:length-1]
}

func getSessionIDIndex(sessionID int) int {
	playerMux.Lock()
	defer playerMux.Unlock()

	for i, val := range currentPlayers {
		if val.sessionID == sessionID {
			return i
		}
	}
	return -1
}

func getPlayerIDIndex(id int) int {
	playerMux.Lock()
	defer playerMux.Unlock()

	for i, val := range currentPlayers {
		if val.id == id {
			return i
		}
	}
	return -1
}

func getJoinIndex(id int) int {
	playerMux.Lock()
	defer playerMux.Unlock()

	for i, val := range joinList {
		if val.id == id {
			return i
		}
	}
	return -1
}

func newIDPair() (int, int) {
	index := 0
	var sessionID int
	for index != -1 {
		sessionID = int(rand.Int())
		index = getSessionIDIndex(sessionID)
	}

	var id int
	index = 0
	for index != -1 {
		id = int(int16(rand.Int()))
		index = getPlayerIDIndex(id)
	}

	return sessionID, id
}

