package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func f1() {
	c1 := make(chan string)
	c2 := make(chan string)
	go func() {
		// time.Sleep(1 * time.Second)
		c1 <- "one"
	}()

	go func() {
		// time.Sleep(2 * time.Second)
		c2 <- "two"
	}()

	for {
		select {
		case m1 := <-c1:
			fmt.Println("received", m1)
		case ms := <-c2:
			fmt.Println("received", ms)
		default:
			fmt.Println("no message received")
		}
	}

}

type Book struct {
	title   string
	author  string
	subject string
	book_id int
}

func printBook(book *Book) {
	fmt.Println(book.title)
	fmt.Println(book.author)
	fmt.Println(book.subject)
	fmt.Println(book.book_id)
}

func main() {
	book1 := Book{title: "Go Programming", author: "Kevin Jin", subject: "Go Programming", book_id: 64}
	fmt.Println(book1)
	printBook(&book1)

	//get absolute path
	absPath, _ := filepath.Abs("./")
	fmt.Println(absPath)

	//get relative path
	relPath, _ := filepath.Rel("./", absPath)
	fmt.Println(relPath)

	filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})

}
