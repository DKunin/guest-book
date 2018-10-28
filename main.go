package main

import (
	"encoding/json"
	"flag"
	"github.com/boltdb/bolt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var (
	addr = flag.String("addr", "127.0.0.1:9898", "addr to bind to")
)

type Server struct {
	Pages map[string]func(io.Writer) error
	mu    sync.Mutex
	Posts []Post
	db    *bolt.DB
}

const index = `
<!DOCTYPE html>
<html>
	<head>
		<title>{{.Title}}</title>
		<link rel="stylesheet" href="/static/css/style.css" />
	</head>
	<body>
        <main>
		{{range .Posts}}
			<article>
				<strong>{{ .Author }}</strong></div><p>{{ .Message}}</p>
			</article>
		{{else}}
			<div><strong>no posts</strong></div>
		{{end}}
		</main>
		<footer>
		<form method="POST" action="/">
			<div><input type="text" name="name" placeholder="Имя"/></div>
			<div><textarea name="text" placeholder="Сообщение"></textarea></div>
			<button>отправить</button>
		</form>
		</footer>
		<script src="/static/js/script.js"></script>
	</body>
</html>`

type Index struct {
	Title string
	Posts []Post
}

type Post struct {
	Author  string
	Message string
}

func (s *Server) GetPosts() []Post {
	s.mu.Lock()
	defer s.mu.Unlock()

	var newPosts []Post
	// TODO: fix initial request of the bucket
	s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Posts"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			post := Post{}
			err := json.Unmarshal(v, &post)
			if err != nil {
				log.Fatal(err)
			}
			newPosts = append(newPosts, post)
		}

		return nil
	})

	return newPosts
}

func (s *Server) GetPostsJSON() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	var newPosts []Post

	s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Posts"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			post := Post{}
			err := json.Unmarshal(v, &post)
			if err != nil {
				log.Fatal(err)
			}
			if post.Author != string("") {
				newPosts = append(newPosts, post)
			}
		}

		return nil
	})

	j, _ := json.Marshal(newPosts)

	return j

}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.Printf("recieved")

	if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			res.WriteHeader(500)
			log.Printf("error with form %s", err)
			return
		}

		post := Post{
			Author:  req.PostFormValue("name"),
			Message: req.PostFormValue("text"),
		}

		s.mu.Lock()

		s.db.Update(func(tx *bolt.Tx) error {
			b, _ := tx.CreateBucketIfNotExists([]byte("Posts"))
			id, _ := b.NextSequence()
			j, _ := json.Marshal(post)
			log.Printf("json: %s", j)
			err := b.Put([]byte(strconv.Itoa(int(id))), j)
			if err != nil {
				log.Printf("broke wrote to db %v", err)
			}
			return err
		})

		s.Posts = append(s.Posts, post)
		s.mu.Unlock()
		res.WriteHeader(200)
		return
	}

	fn, ok := s.Pages[req.URL.Path]

	if !ok {
		res.WriteHeader(404)
		return
	}
	if err := fn(res); err != nil {
		res.WriteHeader(500)
		return
	}

}

func main() {
	flag.Parse()
	tmpl, err := template.New("index").Parse(index)
	if err != nil {
		log.Fatal(err)
	}

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var s *Server
	s = &Server{
		db: db,
		Pages: map[string]func(io.Writer) error{
			"/": func(w io.Writer) error {
				return tmpl.Execute(w, Index{
					Title: "My cool guest book",
					Posts: s.GetPosts(),
				})
			},
			"/posts": func(w io.Writer) error {
				w.Write([]byte(s.GetPostsJSON()))
				return nil
			},
		},
	}

	fs := http.FileServer(http.Dir("./"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(*addr, nil))

}
