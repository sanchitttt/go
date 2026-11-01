package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Item struct{
	Title string 
	Body string 
}


// func (t *T) MethodName(args *Args, reply *Reply) error


type API struct{
	database []Item
};

func (a *API) GetByName(title string, reply *Item) error {
	var item Item 

	for _, val := range a.database {
		if val.Title == title {
			*reply = item;
			break
		}
	}

	return errors.New("Item could not be fond")	
}

// func GetByName(title string) Item{
// 	var item Item 

// 	for _, val := range database {
// 		if val.title == title {
// 			item = val;
// 		}
// 	}

// 	return item;
// }

// func AddItem(item Item) Item{
// 	database = append(database, item)
// 	return item
// }

func (a *API) AddItem(item Item, reply *Item) error{
	a.database = append(a.database, item)
	*reply = item;
	return nil;
}


func (a *API) EditItem(item Item, reply *Item) error{
	for idx, val:= range a.database{
		if item.Title == val.Title {
			a.database[idx] = item;
			*reply = item;
			return nil;
		}
	}

	return errors.New("Item could not be found")
}

// func EditItem(title string, item Item) Item{
// 	var updated Item 

// 	for idx, val:= range database{
// 		if title == val.title {
// 			database[idx] = item;
// 			updated = item;
// 		}
// 	}

// 	return updated
// }

// func DeleteItem(item Item) Item{
// 	var del Item 

// 	for idx, val := range database{
// 		if val.title == item.title && val.body == item.body {
// 			database = append(database[:idx], database[idx+1:]...)
// 			del = item
// 			break
// 		}
// 	}

// 	return del;
// }

func (a *API) GetDB(title string, reply *[]Item) error {
	*reply = a.database 
	return nil
}

func (a *API) DeleteItem(item Item, reply *Item) error{
	for idx, val := range a.database{
		if val.Title == item.Title && val.Body == item.Body {
			a.database = append(a.database[:idx], a.database[idx+1:]...)
			*reply = item
			return nil
		}
	}

	return errors.New("Could not find the item")
}

const PORT = "4040";

func main(){
	var api = new (API);
	fmt.Println("Initial database: ", api.database);
	err := rpc.Register(api)

	if err != nil {
		log.Fatal("Error registering API", err)
	}

	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", ":" + PORT)

	if err != nil {
		log.Fatal("Listener error", err)
	}

	log.Printf("Serving RPC on Port: %s", PORT);

	err = http.Serve(listener, nil)

	if err != nil{
		log.Fatal("Error Serving", err );
	}
		


	// a := Item{"first", "a test item"};
	// b := Item{"second", "b test item"};
	// c := Item{"third", "c test item"};

	// AddItem(a)
	// AddItem(b)
	// AddItem(c)

	// fmt.Println("Database (2): ", database)
	
	// DeleteItem(b)
	
	// fmt.Println("Database (3): ", database)

	// EditItem("third", Item{"fourth", "a new item"})

	// fmt.Println("Database (4): ", database)

	// x := GetByName("fourth")
	// y := GetByName("first")

	// fmt.Println(x, y)
}