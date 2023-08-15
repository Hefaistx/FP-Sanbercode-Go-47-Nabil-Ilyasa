package main

import (
	c "final-project/controller"
	d "final-project/db"
	"net/http"

	_ "final-project/docs"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	//User
	router.POST("/user", c.Register)
	router.POST("/user/login", c.Login)
	router.POST("/user/logout", c.Logout)
	router.GET("/user", c.GetUser)
	router.GET("/user/detail/:id", c.GetUserDetail)
	router.POST("/user/update/:id", c.UpdateUser)
	router.DELETE("/user/:id", c.DeleteUser)
	//Role
	router.POST("/role", c.CreateRole)
	router.DELETE("/role/:id", c.DeleteRole)
	router.GET("/roles", c.GetRole)
	//Game
	router.POST("/game", c.AddGame)
	router.GET("/games", c.GetGames)
	router.GET("/game-detail/:id", c.GetGameDetail)
	router.POST("/game-update/:id", c.UpdateGame)
	router.DELETE("/game/:id", c.DeleteGame)
	//Review
	router.POST("/game/review", c.AddReview)
	router.GET("/game/reviews", c.GetReview)
	router.POST("/review/:id", c.UpdateReview)
	router.DELETE("/review/:uid", c.DeleteReview)
	//Wishlist
	router.POST("/game-wish", c.AddWish)
	router.GET("/game-wish", c.GetWish)
	router.DELETE("/game-wish/delete/:id", c.DeleteWish)
	//Root
	router.GET("/", d.RootHandler)
	// Swagger UI files
	router.ServeFiles("/swagger/*filepath", http.Dir("./docs"))
	http.ListenAndServe(":8080", router)
}
