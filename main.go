package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Article struct {
	ID      uint32
	Title   string
	Name    string
	Content string
}

var Articles []Article

func first(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/main_page.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	file, err := os.Open("articles.csv")
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Ошибка при чтении файла:", err)
		return
	}

	Articles = []Article{}
	for _, record := range records[1:] {
		id, _ := strconv.Atoi(record[0])
		var art Article
		art.ID = uint32(id)
		art.Title = record[1]
		art.Name = record[2]
		art.Content = record[3]
		Articles = append(Articles, art)
	}

	t.ExecuteTemplate(w, "main", Articles)
}

func create(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		author := r.FormValue("author")
		content := r.FormValue("content")

		if title == "" || author == "" || content == "" {
			http.Error(w, "Все поля обязательны", http.StatusBadRequest)
			return
		}

		log.Println(title)
		log.Println(author)
		log.Println(content)

		// db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/articles")
		// if err != nil {
		// 	panic(err)
		// } else {
		// 	log.Println("Open")
		// }
		// defer db.Close()

		// _, err = db.Exec(fmt.Sprintf("INSERT INTO `articles` (`title`, `author`, `content`) VALUES ('%s', '%s', '%s')", title, author, content))
		// if err != nil {
		// 	panic(err)
		// } else {
		// 	log.Println("Insert")
		// }

		newID := 1
		if len(Articles) > 0 {
			newID = len(Articles) + 1
		}

		Articles = append(Articles, Article{
			ID:      uint32(newID),
			Title:   title,
			Name:    author,
			Content: content,
		})

		file, err := os.OpenFile("articles.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("Ошибка при открытии файла:", err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		fileInfo, err := file.Stat()
		if err != nil {
			fmt.Println("Ошибка при получении информации о файле:", err)
			return
		}

		if fileInfo.Size() == 0 {
			err = writer.Write([]string{"ID", "Имя", "Автор", "Содержание"})
			if err != nil {
				fmt.Println("Ошибка при записи заголовков:", err)
				return
			}
		}

		err = writer.Write([]string{string(newID), title, author, content})
		if err != nil {
			fmt.Println("Ошибка при записи статьи:", err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("templates/create.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	t.ExecuteTemplate(w, "create", nil)
}

func show_post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Title: %v\n", vars["title"])
}

func deleteArticleByID(articleID int) error {
	// Читаем существующие статьи из CSV
	file, err := os.Open("articles.csv")
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.ReadAll()
	if err != nil {
		fmt.Println("Ошибка при чтении файла:", err)
		return nil
	}

	// Фильтруем статьи, исключая ту, которую нужно удалить
	var updatedArticles []Article
	for _, article := range Articles {
		if article.ID != uint32(articleID) {
			updatedArticles = append(updatedArticles, article)
		}
	}

	// Записываем обновленный список статей обратно в CSV
	file, err = os.Create("articles.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записываем заголовок
	writer.Write([]string{"id", "title", "author", "content"})
	for _, article := range updatedArticles {
		record := []string{string(article.ID), article.Title, article.Name, article.Content}
		writer.Write(record)
	}

	return nil
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleID := vars["id"] // Получаем ID статьи из URL

	intarticleID, err := strconv.Atoi(articleID)
	err = deleteArticleByID(intarticleID)
	if err != nil {
		http.Error(w, "Ошибка при удалении статьи", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Успешно удалено
}

func handleFunc() {
	rtr := mux.NewRouter()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	rtr.HandleFunc("/", first).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET", "POST")
	// http.HandleFunc("/save_article", save_article)
	rtr.HandleFunc("/post/{title}", show_post).Methods("GET")
	rtr.HandleFunc("/delete/{id}", deleteArticle).Methods("DELETE")
	http.ListenAndServe(":8081", rtr)
}

func main() {
	handleFunc()
}

// func save_article(w http.ResponseWriter, r *http.Request) {
// 	log.Println("Save_article")
// 	title := r.FormValue("title")
// 	author := r.FormValue("author")
// 	content := r.FormValue("content")

// 	if title == "" || author == "" || content == "" {
// 		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
// 		return
// 	}

// 	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/articles")
// 	if err != nil {
// 		panic(err)
// 	} else {
// 		log.Println("Open")
// 	}
// 	defer db.Close()

// 	_, err = db.Exec(fmt.Sprintf("INSERT INTO `articles` (`title`, `author`, `content`) VALUES ('%s', '%s', '%s')", title, author, content))
// 	if err != nil {
// 		panic(err)
// 	} else {
// 		log.Println("Insert")
// 	}

// 	// Articles = append(Articles, Article{
// 	// 	Title:   title,
// 	// 	Name:    author,
// 	// 	Content: content,
// 	// })

// 	// file, err := os.OpenFile("articles.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
// 	// if err != nil {
// 	// 	fmt.Println("Ошибка при открытии файла:", err)
// 	// 	return
// 	// }
// 	// defer file.Close()

// 	// writer := csv.NewWriter(file)
// 	// defer writer.Flush()

// 	// fileInfo, err := file.Stat()
// 	// if err != nil {
// 	// 	fmt.Println("Ошибка при получении информации о файле:", err)
// 	// 	return
// 	// }

// 	// if fileInfo.Size() == 0 {
// 	// 	err = writer.Write([]string{"Имя", "Автор", "Содержание"})
// 	// 	if err != nil {
// 	// 		fmt.Println("Ошибка при записи заголовков:", err)
// 	// 		return
// 	// 	}
// 	// }

// 	// for _, article := range Articles {
// 	// 	err := writer.Write([]string{article.Title, article.Name, article.Content})
// 	// 	if err != nil {
// 	// 		fmt.Println("Ошибка при записи статьи:", err)
// 	// 		return
// 	// 	}
// 	// }

// 	http.Redirect(w, r, "/", http.StatusSeeOther)
// }
