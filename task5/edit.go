package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"net/http/cgi"
)

type EditPageData struct {
	PageData
	Login string
}

var editTemplate = template.Must(template.New("edit").Parse(editHTML))

func renderEdit(w http.ResponseWriter, data EditPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data.Languages = allLanguages
	editTemplate.Execute(w, data)
}

func requireAuth(w http.ResponseWriter, r *http.Request) (JWTPayload, bool) {
	payload, ok := getJWTFromCookie(r)
	if !ok {
		http.Redirect(w, r, "login.cgi", http.StatusFound)
		return JWTPayload{}, false
	}
	return payload, true
}

func runEdit(db *sql.DB) {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleEditGet(w, r, db)
		case http.MethodPost:
			handleEditPost(w, r, db)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func handleEditGet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	payload, ok := requireAuth(w, r)
	if !ok {
		return
	}

	pageData := loadFromCookies(w, r)

	if len(pageData.Errors) == 0 {
		appData, err := getApplicationByID(db, payload.ApplicationID)
		if err != nil {
			log.Println("getApplicarionByID:", err)
			http.Error(w, "Failed to load application data", http.StatusInternalServerError)
			return
		}
		pageData.Values = appData
	}

	renderEdit(w, EditPageData{
		PageData: pageData,
		Login:    payload.Login,
	})
}

func handleEditPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	payload, ok := requireAuth(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	data, errors := validate(r)

	if len(errors) > 0 {
		saveErrorsToCookie(w, data, errors)
		http.Redirect(w, r, "edit.cgi", http.StatusNotFound)
		return
	}

	if err := updateApplication(db, payload.ApplicationID, data); err != nil {
		log.Println("updateApplication:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	saveSuccessToCookie(w, data)
	http.Redirect(w, r, "edit.cgi", http.StatusFound)
}

const editHTML = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Редактирование анкеты</title>
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
            padding: 2rem;
            background-image: radial-gradient(circle at 10% 30%, rgba(140, 100, 70, 0.04) 2%, transparent 2.5%);
            background-size: 28px 28px;
        }

        /* верхняя панель — как скромная плашка */
        .topbar {
            max-width: 720px;
            margin: 0 auto 1.2rem auto;
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: rgba(250, 245, 235, 0.8);
            border-left: 1px solid #dacbb8;
            border-right: 1px solid #dacbb8;
            padding: 0.6rem 1.5rem;
        }

        .topbar-user {
            font-size: 0.8rem;
            letter-spacing: 1px;
            color: #8b694c;
        }

        .topbar-user strong {
            color: #6a4e2e;
            font-weight: 500;
        }

        .topbar a {
            font-size: 0.75rem;
            letter-spacing: 1px;
            color: #8b694c;
            text-decoration: none;
            border: 0.5px solid #dacbb8;
            padding: 0.2rem 0.8rem;
            transition: all 0.2s;
        }

        .topbar a:hover {
            background: #e8ddd0;
            color: #4a3924;
        }

        /* основной свиток */
        .scroll-form {
            max-width: 720px;
            width: 100%;
            margin: 0 auto;
            background: rgba(250, 245, 235, 0.92);
            border-left: 1px solid #dacbb8;
            border-right: 1px solid #dacbb8;
            padding: 2rem 2rem 2.2rem;
            box-shadow: 0 20px 30px -15px rgba(0,0,0,0.05), inset 0 0 0 1px rgba(255,250,240,0.5);
        }

        h1 {
            font-size: 1.7rem;
            font-weight: 400;
            letter-spacing: 3px;
            color: #6a4e2e;
            text-align: center;
            margin-bottom: 1.8rem;
            padding-bottom: 0.6rem;
            border-bottom: 0.5px solid #e2d5c4;
            display: inline-block;
            width: auto;
            margin-left: auto;
            margin-right: auto;
        }

        h1::before {
            content: "・";
            margin-right: 10px;
            color: #b87c4f;
        }
        h1::after {
            content: "・";
            margin-left: 10px;
            color: #b87c4f;
        }

        .field {
            margin-bottom: 1.4rem;
        }

        .field > label {
            display: block;
            font-size: 0.8rem;
            letter-spacing: 2px;
            color: #8b694c;
            margin-bottom: 0.4rem;
            font-weight: 400;
            text-transform: uppercase;
        }

        input[type="text"],
        input[type="tel"],
        input[type="email"],
        input[type="date"],
        select,
        textarea {
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

        input:focus,
        select:focus,
        textarea:focus {
            border-color: #b28b6f;
            background: #ffffff;
        }

        .field-error input,
        .field-error select,
        .field-error textarea {
            border-color: #c9826b;
            background: #fffaf5;
        }

        .error-msg {
            font-size: 0.7rem;
            color: #b16245;
            margin-top: 0.3rem;
            margin-left: 0.3rem;
            display: flex;
            align-items: center;
            gap: 4px;
        }

        .error-msg::before {
            content: "・";
            font-size: 0.8rem;
            color: #b87c4f;
        }

        textarea {
            min-height: 90px;
            resize: vertical;
        }

        select[multiple] {
            height: 150px;
            background: #fefaf5;
        }

        select[multiple] option {
            padding: 0.3rem 0.5rem;
            margin: 2px 0;
        }

        select[multiple] option:checked {
            background: #e8ddd0 linear-gradient(0deg, #dccbb8 0%, #dccbb8 100%);
            color: #4a2e1a;
        }

        .radio-group {
            display: flex;
            gap: 1.8rem;
            flex-wrap: wrap;
            align-items: center;
            margin-top: 0.3rem;
        }

        .radio-group label,
        .checkbox-label {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            font-weight: 400;
            color: #6a4e2e;
            font-size: 0.9rem;
            cursor: pointer;
        }

        input[type="radio"],
        input[type="checkbox"] {
            accent-color: #b87c4f;
            width: auto;
            margin: 0;
        }

        .btn {
            width: 100%;
            background: transparent;
            border: 1px solid #b28b6f;
            padding: 0.8rem;
            font-size: 0.9rem;
            letter-spacing: 3px;
            color: #6a4e2e;
            font-family: inherit;
            cursor: pointer;
            transition: all 0.2s;
            margin-top: 0.6rem;
            background: rgba(210, 180, 140, 0.05);
            text-transform: uppercase;
        }

        .btn:hover {
            background: #e8ddd0;
            border-color: #8b694c;
            color: #3e2a1a;
            letter-spacing: 4px;
        }

        .success-banner {
            background: #efe4d8;
            border: 1px solid #dacbb8;
            padding: 0.8rem 1rem;
            color: #6a4e2e;
            font-size: 0.85rem;
            margin-bottom: 1.8rem;
            text-align: center;
            letter-spacing: 1px;
        }

        .hanko {
            text-align: center;
            margin-top: 1.8rem;
            font-size: 0.65rem;
            color: #b28b6f;
            letter-spacing: 2px;
            font-family: monospace;
            border-top: 0.5px solid #e2d5c4;
            padding-top: 1.2rem;
            width: 80%;
            margin-left: auto;
            margin-right: auto;
        }

        @media (max-width: 600px) {
            body {
                padding: 1rem;
            }
            .scroll-form {
                padding: 1.2rem;
            }
            h1 {
                font-size: 1.4rem;
                letter-spacing: 2px;
            }
            .radio-group {
                flex-direction: column;
                gap: 0.5rem;
            }
            .topbar {
                flex-direction: column;
                gap: 0.5rem;
                text-align: center;
            }
        }
    </style>
</head>
<body>

<div class="topbar">
    <span class="topbar-user">Вы вошли как <strong>{{.Login}}</strong></span>
    <a href="logout.cgi">Выйти</a>
</div>

<div class="scroll-form">
    <h1>✎ Редактирование анкеты</h1>

    {{if .Success}}
    <div class="success-banner">✓ Данные успешно обновлены</div>
    {{end}}

    <form action="edit.cgi" method="POST">

        <div class="field {{if index .Errors "name"}}field-error{{end}}">
            <label>ФИО</label>
            <input type="text" name="name" value="{{.Values.Name}}">
            {{if index .Errors "name"}}
                <div class="error-msg">{{index .Errors "name"}}</div>
            {{end}}
        </div>

        <div class="field {{if index .Errors "phone"}}field-error{{end}}">
            <label>Телефон</label>
            <input type="tel" name="phone" value="{{.Values.Phone}}">
            {{if index .Errors "phone"}}
                <div class="error-msg">{{index .Errors "phone"}}</div>
            {{end}}
        </div>

        <div class="field {{if index .Errors "email"}}field-error{{end}}">
            <label>Email</label>
            <input type="email" name="email" value="{{.Values.Email}}">
            {{if index .Errors "email"}}
                <div class="error-msg">{{index .Errors "email"}}</div>
            {{end}}
        </div>

        <div class="field {{if index .Errors "birthdate"}}field-error{{end}}">
            <label>Дата рождения</label>
            <input type="date" name="birthdate" value="{{.Values.Birthdate}}">
            {{if index .Errors "birthdate"}}
                <div class="error-msg">{{index .Errors "birthdate"}}</div>
            {{end}}
        </div>

        <div class="field {{if index .Errors "gender"}}field-error{{end}}">
            <label>Пол</label>
            <div class="radio-group">
                <label>
                    <input type="radio" name="gender" value="male"
                        {{if eq .Values.Gender "male"}}checked{{end}}> Мужской
                </label>
                <label>
                    <input type="radio" name="gender" value="female"
                        {{if eq .Values.Gender "female"}}checked{{end}}> Женский
                </label>
            </div>
            {{if index .Errors "gender"}}
                <div class="error-msg">{{index .Errors "gender"}}</div>
            {{end}}
        </div>

        <div class="field {{if index .Errors "languages"}}field-error{{end}}">
            <label>Любимый язык программирования</label>
            <select name="languages[]" multiple>
                {{range .Languages}}
                <option value="{{.ID}}"
                    {{if $.IsSelectedLang .ID}}selected{{end}}>{{.Name}}</option>
                {{end}}
            </select>
            {{if index .Errors "languages"}}
                <div class="error-msg">{{index .Errors "languages"}}</div>
            {{end}}
        </div>

        <div class="field {{if index .Errors "bio"}}field-error{{end}}">
            <label>Биография</label>
            <textarea name="bio">{{.Values.Bio}}</textarea>
            {{if index .Errors "bio"}}
                <div class="error-msg">{{index .Errors "bio"}}</div>
            {{end}}
        </div>

        <div class="field {{if index .Errors "contract"}}field-error{{end}}">
            <label class="checkbox-label">
                <input type="checkbox" name="contract"
                    {{if .Values.Contract}}checked{{end}}> С контрактом ознакомлен(а)
            </label>
            {{if index .Errors "contract"}}
                <div class="error-msg">{{index .Errors "contract"}}</div>
            {{end}}
        </div>

        <button type="submit" class="btn">Сохранить изменения</button>
    </form>

    <div class="hanko">⦿ 礼</div>
</div>
</body>
</html>
`
