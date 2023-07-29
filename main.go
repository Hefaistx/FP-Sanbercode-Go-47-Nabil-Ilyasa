package main

import (
	c "final-project/controller"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	//User
	router.POST("/user", c.Register)
	router.POST("/user/login", c.Login)
	router.DELETE("/user/:id", c.DeleteUser)
	//Role
	router.POST("/role", c.CreateRole)
	router.DELETE("/role/:id", c.DeleteRole)
	//Game
	router.POST("/game", c.AddGame)
	//Review
	router.POST("/game/review", c.AddReview)
	//Wishlist
	router.POST("/game/wish", c.AddWish)
	http.ListenAndServe(":8080", router)
}
