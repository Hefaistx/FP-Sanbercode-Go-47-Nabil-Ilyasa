package main

import (
	c "final-project/controller"
	d "final-project/db"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	//User
	router.POST("/user", c.Register)
	router.POST("/user/login", c.Login)
	router.GET("/user", c.GetUser)
	router.GET("/user/:id", c.GetUserDetail)
	router.POST("/user/:id", c.UpdateUser)
	router.DELETE("/user/:id", c.DeleteUser)
	//Role
	router.POST("/role", c.CreateRole)
	router.DELETE("/role/:id", c.DeleteRole)
	router.GET("/roles", c.GetRole)
	//Game
	router.POST("/game", c.AddGame)
	router.GET("/games", c.GetGames)
	router.GET("/game/:id", c.GetGameDetail)
	router.POST("/game/:id", c.UpdateGame)
	router.DELETE("/game/:id", c.DeleteGame)
	//Review
	router.POST("/game/review", c.AddReview)
	router.GET("/game/reviews", c.GetReview)
	router.POST("/game/review:id", c.UpdateReview)
	router.DELETE("/game/review/:id", c.DeleteReview)
	//Wishlist
	router.POST("/game/wish", c.AddWish)
	router.GET("/game/wish", c.GetWish)
	router.DELETE("/game/wish", c.DeleteWish)
	http.ListenAndServe(":8080", router)
	//Root
	http.HandleFunc("/", d.RootHandler)
}
