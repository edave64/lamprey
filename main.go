package main

import (
	"lamprey/core"
	"lamprey/core/db"
	"lamprey/core/deployer"
	"net"
	"net/http"
	"net/http/fcgi"
)

type App struct {
	PageManager db.PageManager
	Config      core.Config
	Deployer    deployer.Deployer
}

func main() {
	a := core.ReadConfig("config.toml")
	pm, err := db.NewPageManager("db.sqlite")
	if err != nil {
		panic(err)
	}
	app := App{
		PageManager: *pm,
		Config:      a,
	}

	// start server
	if a.Http != nil {
		httpServer(*a.Http, app)
	}

	if a.FastCgi != nil {
		fastCgiServer(*a.FastCgi, app)
	}
}

func httpServer(config core.HttpConfig, app App) {
	println("Starting http server on " + config.Address)

	if app.Config.DeployToFolder != nil {
		http.ListenAndServe(config.Address, getHttpHandler("/lamprey", app))
	} else {
		http.ListenAndServe(config.Address, getHttpHandler("", app))
	}
}

func fastCgiServer(config core.FastCgiConfig, app App) {
	println("Starting fastcgi server on " + config.Address)

	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		println("Unable to bind to address:" + config.Address)
		panic(err)
	}
	err = fcgi.Serve(listener, getHttpHandler("", app))
	if err != nil {
		println("Unable to start fastcgi server")
		panic(err)
	}

}

func getHttpHandler(prefix string, app App) http.Handler {
	a := http.ServeMux{}
	a.HandleFunc(prefix+"/admin", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Only admins should be able to access this page"))
		w.Write([]byte(r.RequestURI))
	})
	a.HandleFunc("GET "+prefix+"/edit/article/{objectid}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Editing article for " + r.PathValue("objectid")))
		w.Write([]byte(r.RequestURI))
	})
	a.HandleFunc("POST "+prefix+"/edit/article/{objectid}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Saving article for " + r.PathValue("objectid")))
		w.Write([]byte(r.RequestURI))

		page, err := app.PageManager.GetPage(r.PathValue("objectid"))
		if err != nil {
			w.Write([]byte("Unable to get page"))
			panic(err)
		}
		page.Content = r.FormValue("content")
		err = app.PageManager.UpdatePage(page.ID, page.Title, page.Content, page.Data)
		if err != nil {
			w.Write([]byte("Unable to update page"))
			panic(err)
		}
		err = app.Deployer.DeployArticle(*page)
		if err != nil {
			w.Write([]byte("Unable to deploy article"))
			panic(err)
		}
	})
	a.HandleFunc("GET "+prefix+"/data/{objectid}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Editing data for " + r.PathValue("objectid")))
		w.Write([]byte(r.RequestURI))
	})
	a.HandleFunc("POST "+prefix+"/data/{objectid}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Saving data for " + r.PathValue("objectid")))
		w.Write([]byte(r.RequestURI))

		page, err := app.PageManager.GetPage(r.PathValue("objectid"))
		if err != nil {
			w.Write([]byte("Unable to get page"))
			panic(err)
		}
		page.Content = r.FormValue("content")
		err = app.PageManager.UpdatePage(page.ID, page.Title, page.Content, page.Data)
		if err != nil {
			w.Write([]byte("Unable to update page"))
			panic(err)
		}
		err = app.Deployer.DeployData(*page)
		if err != nil {
			w.Write([]byte("Unable to deploy article"))
			panic(err)
		}
	})
	a.HandleFunc("GET "+prefix+"/new", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Creating new entry"))
		w.Write([]byte(r.RequestURI))
	})
	a.HandleFunc("PUT "+prefix+"/new", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Creating new entry"))
		w.Write([]byte(r.RequestURI))
	})
	if app.Config.DeployToFolder != nil {
		println("Staticly serving from folder: " + app.Config.DeployToFolder.Path)
		a.Handle("/", http.FileServer(http.Dir(app.Config.DeployToFolder.Path)))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.ServeHTTP(w, r)
	})
}
