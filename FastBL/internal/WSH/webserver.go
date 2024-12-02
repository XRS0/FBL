package wsh

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// путь к фронтенд-директории
const frontEndPath = "./FrontEnd"

func StartWebServer() {
	// Роутинг для статических файлов
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(filepath.Join(frontEndPath, "assets")))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(frontEndPath, "static")))))

	// Главная страница
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		renderTemplate(w, "index.html")
	})

	// Запуск сервера
	port := ":8080"
	log.Printf("Сервер запущен на http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// renderTemplate рендерит HTML-шаблоны
func renderTemplate(w http.ResponseWriter, tmpl string) {
	tmplPath := filepath.Join(frontEndPath, "templates", tmpl)
	tmplParsed, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		log.Println("Ошибка парсинга шаблона:", err)
		return
	}
	err = tmplParsed.Execute(w, nil)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		log.Println("Ошибка выполнения шаблона:", err)
	}
}
