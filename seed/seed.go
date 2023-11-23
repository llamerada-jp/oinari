/*
 * Copyright 2018 Yuji Ito <llamerada.jp@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package seed

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"path/filepath"

	"net/http"

	"html/template"

	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const (
	SESSION_KEY                = "oinari"
	SECRET_KEY_COOKIE_KEY_PAIR = "cookie_key_pair"

	SECRET_KEY_GITHUB_CLIENT_ID     = "github_client_id"
	SECRET_KEY_GITHUB_CLIENT_SECRET = "github_client_secret"
)

var (
	embed        map[string]string
	templateRoot string

	oauth2Github *oauth2.Config
)

func Init(mux *http.ServeMux, secret map[string]string, templatePath string, withoutSignin bool) error {
	embed = secret
	templateRoot = templatePath

	// setup cookie store
	cookieKeyPair, err := base64.StdEncoding.DecodeString(secret[SECRET_KEY_COOKIE_KEY_PAIR])
	if err != nil {
		return fmt.Errorf("failed to decode cookie key pair: %w", err)
	}
	store := sessions.NewCookieStore(cookieKeyPair)
	store.Options.HttpOnly = true

	// setup oauth
	oauth2Github = &oauth2.Config{
		ClientID:     secret[SECRET_KEY_GITHUB_CLIENT_ID],
		ClientSecret: secret[SECRET_KEY_GITHUB_CLIENT_SECRET],
		Endpoint:     endpoints.GitHub,
		RedirectURL:  "https://localhost:8080/callback_github",
		Scopes:       []string{"user"},
	}

	// setup handlers
	mux.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		// check session
		if !withoutSignin {
			session, err := store.Get(r, SESSION_KEY)
			if err != nil {
				log.Printf("failed to get session: %v", err)
				writeError(w)
				return
			}

			if session.IsNew {
				writePage(w, "signin.html")
				return
			}

			_, ok := session.Values["auth_type"]
			if !ok {
				writePage(w, "signin.html")
				return
			}
		}

		writePage(w, "main.html")
	})

	mux.HandleFunc("/signin_github", func(w http.ResponseWriter, r *http.Request) {
		session := sessions.NewSession(store, SESSION_KEY)

		// create new state string
		binaryData := make([]byte, 32)
		if _, err := rand.Read(binaryData); err != nil {

		}
		state := base64.StdEncoding.EncodeToString(binaryData)
		session.Values["signin_github_state"] = state
		session.Save(r, w)

		url := oauth2Github.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/callback_github", func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, SESSION_KEY)
		if err != nil {
			log.Printf("failed to get session: %v", err)
			writeError(w)
			return
		}

		if session.IsNew {
			http.Redirect(w, r, "https://localhost:8080/index.html", http.StatusTemporaryRedirect)
			return
		}

		// check state
		expectedState, ok := session.Values["signin_github_state"]
		if !ok {
			log.Printf("failed to get state")
			writeError(w)
			return
		}
		if r.URL.Query().Get("state") != expectedState {
			log.Printf("state is invalid")
			writeError(w)
			return
		}

		// exchange code for token & get github user
		code := r.URL.Query().Get("code")
		tok, err := oauth2Github.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("Unable to exchange code for token")
			writeError(w)
			return
		}

		client := github.NewClient(oauth2Github.Client(context.Background(), tok))

		user, _, err := client.Users.Get(context.Background(), "")
		if err != nil {
			log.Printf("Unable to get github user")
			writeError(w)
			return
		}

		// recreate session
		session.Options.MaxAge = -1
		session = sessions.NewSession(store, SESSION_KEY)
		session.Values["auth_type"] = "github"
		session.Values["user"] = *user.Login
		// session.Values["github_token"] = tok
		session.Save(r, w)

		log.Printf("sign-in success GitHub:%s from:%s", *user.Login, r.RemoteAddr)

		http.Redirect(w, r, "./index.html", http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/signout", func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, SESSION_KEY)
		if err != nil {
			log.Printf("failed to get session: %v", err)
			writeError(w)
			return
		}

		if session.IsNew {
			http.Redirect(w, r, "./index.html", http.StatusTemporaryRedirect)
			return
		}

		authType := session.Values["auth_type"]
		switch authType {
		case "github":
			// TODO: should I revoke github token?
		}

		// destroy session
		session.Options.MaxAge = -1
		session.Save(r, w)

		http.Redirect(w, r, "./index.html", http.StatusTemporaryRedirect)
	})

	return nil
}

func writeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	writePage(w, "error.html")
}

func writePage(w http.ResponseWriter, file string) {
	tpl, err := template.ParseFiles(filepath.Join(templateRoot, file))
	if err != nil {
		log.Fatalf("failed to read template file: %v", err)
	}

	if err = tpl.Execute(w, embed); err != nil {
		log.Fatalf("failed to generate html: %v", err)
	}
}
