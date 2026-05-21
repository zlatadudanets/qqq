package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"net/http/cgi"
)

type LoginPageData struct {
	Error string
}

var loginTemplate = template.Must(template.New("login").Parse(loginHTML))

func renderLogin(w http.ResponseWriter, data LoginPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	loginTemplate.Execute(w, data)
}

func runLogin(db *sql.DB) {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLoginGet(w, r)
		case http.MethodPost:
			handleLoginPost(w, r, db)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func handleLoginGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := getJWTFromCookie(r); ok {
		http.Redirect(w, r, "edit.cgi", http.StatusFound)
		return
	}
	renderLogin(w, LoginPageData{})
}

func handleLoginPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	login := r.FormValue("login")
	password := r.FormValue("password")

	if login == "" || password == "" {
		renderLogin(w, LoginPageData{Error: "Login and password cannot be empty"})
		return
	}

	creds, err := findCredentialsByLogin(db, login)
	if err != nil {
		log.Println("findCredentialsByLogin:", err)
		renderLogin(w, LoginPageData{Error: "Invalid login or password"})
		return
	}

	if !checkPassword(password, creds.PasswordHash) {
		renderLogin(w, LoginPageData{Error: "Invalid login or password"})
		return
	}

	token, err := generateJWT(creds.ApplicationID, login)
	if err != nil {
		log.Println("generateJWT:", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	setJWTCookie(w, token)
	http.Redirect(w, r, "edit.cgi", http.StatusFound)
}

const loginHTML = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Вход · восточный стиль</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            background: #f3efe7;
            font-family: 'Noto Serif SC', 'Noto Serif JP', 'Times New Roman', '游明朝', 'Yu Mincho', Georgia, serif;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 2rem;
            background-image: radial-gradient(circle at 10% 30%, rgba(140, 100, 70, 0.04) 2%, transparent 2.5%);
            background-size: 28px 28px;
        }

        /* свиток */
        .scroll-card {
            max-width: 440px;
            width: 100%;
            background: rgba(250, 245, 235, 0.92);
            border-left: 1px solid #dacbb8;
            border-right: 1px solid #dacbb8;
            padding: 2rem 1.8rem 2.2rem;
            box-shadow: 0 20px 30px -15px rgba(0,0,0,0.05), inset 0 0 0 1px rgba(255,250,240,0.5);
            text-align: center;
        }

        /* эмблема */
        .emblem {
            font-size: 2.8rem;
            margin-bottom: 0.8rem;
            display: inline-block;
            color: #8b694c;
        }

        h1 {
            font-size: 1.7rem;
            font-weight: 400;
            letter-spacing: 3px;
            color: #6a4e2e;
            margin-bottom: 1.5rem;
            padding-bottom: 0.4rem;
            border-bottom: 0.5px solid #e2d5c4;
            display: inline-block;
        }

        h1::before {
            content: "・";
            margin-right: 8px;
            color: #b87c4f;
        }
        h1::after {
            content: "・";
            margin-left: 8px;
            color: #b87c4f;
        }

        .field {
            margin-bottom: 1.2rem;
            text-align: left;
        }

        label {
            display: block;
            font-size: 0.75rem;
            letter-spacing: 2px;
            text-transform: uppercase;
            color: #8b694c;
            margin-bottom: 0.3rem;
        }

        input {
            width: 100%;
            padding: 0.7rem 0.8rem;
            background: #fefaf5;
            border: 1px solid #e2d5c4;
            font-family: inherit;
            font-size: 0.9rem;
            color: #4a3924;
            transition: all 0.2s;
            outline: none;
            border-radius: 0;
        }

        input:focus {
            border-color: #b28b6f;
            background: #ffffff;
        }

        .error-banner {
            background: #fef0e8;
            border: 1px solid #e2c8b8;
            padding: 0.6rem 1rem;
            color: #b16245;
            font-size: 0.8rem;
            margin-bottom: 1.5rem;
            letter-spacing: 1px;
        }

        .btn {
            width: 100%;
            background: transparent;
            border: 1px solid #b28b6f;
            padding: 0.7rem;
            font-size: 0.85rem;
            letter-spacing: 3px;
            color: #6a4e2e;
            font-family: inherit;
            cursor: pointer;
            transition: all 0.2s;
            margin-top: 0.5rem;
            text-transform: uppercase;
        }

        .btn:hover {
            background: #e8ddd0;
            border-color: #8b694c;
            letter-spacing: 4px;
        }

        .links {
            margin-top: 1.5rem;
            font-size: 0.75rem;
            letter-spacing: 1px;
        }

        .links a {
            color: #9e7b5e;
            text-decoration: none;
            border-bottom: 0.5px dotted #dacbb8;
            transition: color 0.2s;
        }

        .links a:hover {
            color: #6a4e2e;
            border-bottom-color: #8b694c;
        }

        .hanko {
            margin-top: 1.8rem;
            font-size: 0.65rem;
            color: #b28b6f;
            letter-spacing: 2px;
            font-family: monospace;
            border-top: 0.5px solid #e2d5c4;
            padding-top: 1rem;
            width: 60%;
            margin-left: auto;
            margin-right: auto;
        }

        @media (max-width: 500px) {
            body {
                padding: 1rem;
            }
            .scroll-card {
                padding: 1.2rem;
            }
            h1 {
                font-size: 1.4rem;
            }
        }
    </style>
</head>
<body>
<div class="scroll-card">
    <div class="emblem">⛩️</div>
    <h1>Вход</h1>

    {{if .Error}}
    <div class="error-banner">{{.Error}}</div>
    {{end}}

    <form action="login.cgi" method="POST">
        <div class="field">
            <label>Логин</label>
            <input type="text" name="login" autocomplete="username">
        </div>
        <div class="field">
            <label>Пароль</label>
            <input type="password" name="password" autocomplete="current-password">
        </div>
        <button type="submit" class="btn">Войти</button>
    </form>

    <div class="links">
        <a href="form.cgi">← заполнить новую анкету</a>
    </div>

    <div class="hanko">⦿ 礼</div>
</div>
</body>
</html>
`
