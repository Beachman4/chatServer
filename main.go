package main

import (
	"log"
	"net/http"
	"github.com/googollee/go-socket.io"
	"github.com/satyakb/go-socket.io-redis"
	"github.com/rs/cors"
	"math/rand"
	"time"
	"encoding/json"
	"os"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type Conversations struct {
	user1, user2 string
	socket1, socket2 string

}

type ConnectedUsers struct {
	username string
	socket socketio.Socket

}
var users = map[string]ConnectedUsers{}
var server, err = socketio.NewServer(nil)
func waiteSome() {
	for true {
		time.Sleep(5 * time.Second)
		server.BroadcastTo("/", "blah", func() {
			log.Println("sent")
		})
		log.Println(users)
	}
}

func garbageCollecton() {
	for true {
		time.Sleep(2 * time.Second)
		exists := make(map[string]int)
		for key, value := range users {
			_, ok := exists[value.username]
			if ok {
				delete(users, key)
			} else {
				exists[value.username] = 1
			}
		}
	}
}

func checkIfExists(username string) bool {
	_, ok := users[username]
	if ok {
		return false
	}
	return true
}

func sendListOfUsers() {

	for true {
		time.Sleep(5 * time.Second)
		userList := make(map[string]string)
		for _, value := range users {
			userList[value.username] = value.username;
		}
		json.NewEncoder(os.Stdout).Encode(userList)
		for _, u := range users {
			so := u.socket
			so.Emit("userlist", userList)
		}
	}
}

func deleteAllUserByName(socket string) {

}

func main() {

	mux := http.NewServeMux()
	/*mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	})*/


	//conversations := map[string]Conversations{}

	if err != nil {
		log.Fatal(err)
	}

	opts := make(map[string]string)
	server.SetAdaptor(redis.Redis(opts))

	server.On("connection", func (so socketio.Socket) {
		so.On("userinfo", func(msg string) {
			if checkIfExists(msg) {
				users[msg] = ConnectedUsers{
					msg,
					so,
				}
			}
		})

		so.On("messageRequest", func(msg string) {

		})

		so.On("disconnection", func() {
			deleteAllUserByName(so.Id())
		})
	})


	go sendListOfUsers()


	mux.Handle("/socket.io/", server)
	mux.Handle("/", http.FileServer(http.Dir("./asset")))
	handler := cors.Default().Handler(mux)
	go waiteSome()
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowCredentials: true,
	})

	handler = c.Handler(handler)
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", handler))
}
