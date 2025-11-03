package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Task struct {
	ID          int
	Title       string
	Description string
	Reward      string
	Difficulty  string
	Category    string
	Status      string
}

type User struct {
	Username     string  `json:"username"`
	Balance      float64 `json:"balance"`
	Completed    int     `json:"completed"`
	MemberSince  string  `json:"member_since"`
	Level        int     `json:"level"`
	Rank         string  `json:"rank"`
	TotalEarned  float64 `json:"total_earned"`
	SuccessRate  int     `json:"success_rate"`
	CurrentStreak int    `json:"current_streak"`
	Referrals    int     `json:"referrals"`
	IsGuest      bool    `json:"is_guest"`
}

type AppData struct {
	Users      map[string]*User `json:"users"`
	TapBalance float64          `json:"tap_balance"`
}

type VisitorStats struct {
	TotalVisitors   int `json:"total_visitors"`
	UniqueVisitors  int `json:"unique_visitors"`
	OnlineNow       int `json:"online_now"`
}

// Глобальное состояние
var (
	gameMutex  sync.Mutex
	appData    = &AppData{
		Users:      make(map[string]*User),
		TapBalance: 0,
	}
	currentUser = "" // текущий пользователь (пустой если не авторизован)
	dataFile    = "data.json"
	
	// Статистика посещений
	visitorStats = &VisitorStats{
		TotalVisitors:  0,
		UniqueVisitors: 0,
		OnlineNow:      0,
	}
	visitorsMutex    sync.Mutex
	activeSessions   = make(map[string]time.Time) // активные сессии
	uniqueVisitors   = make(map[string]bool)      // уникальные посетители
)

func init() {
	loadData()
}

func loadData() {
	file, err := os.ReadFile(dataFile)
	if err != nil {
		log.Println("Не удалось загрузить данные, создаем новые:", err)
		return
	}
	
	err = json.Unmarshal(file, &appData)
	if err != nil {
		log.Println("Ошибка чтения данных:", err)
	}
}

func saveData() {
	file, err := json.MarshalIndent(appData, "", "  ")
	if err != nil {
		log.Println("Ошибка сериализации данных:", err)
		return
	}
	
	err = os.WriteFile(dataFile, file, 0644)
	if err != nil {
		log.Println("Ошибка сохранения данных:", err)
	}
}

// Функция для вывода статистики в терминал
func printVisitorStats() {
	fmt.Printf("\r📊 Статистика: Всего посещений: %d | Уникальных: %d | Онлайн сейчас: %d", 
		visitorStats.TotalVisitors, visitorStats.UniqueVisitors, visitorStats.OnlineNow)
}

// Функция для отслеживания посетителей
func trackVisitor(r *http.Request) {
	visitorsMutex.Lock()
	defer visitorsMutex.Unlock()
	
	// Получаем IP адрес посетителя
	ip := r.RemoteAddr
	// Для простоты используем IP как идентификатор сессии
	sessionID := ip
	
	// Увеличиваем общее количество посещений
	visitorStats.TotalVisitors++
	
	// Проверяем уникального посетителя
	if !uniqueVisitors[sessionID] {
		uniqueVisitors[sessionID] = true
		visitorStats.UniqueVisitors++
	}
	
	// Обновляем активную сессию
	activeSessions[sessionID] = time.Now()
	
	// Очищаем старые сессии (более 15 минут неактивности)
	now := time.Now()
	for session, lastActivity := range activeSessions {
		if now.Sub(lastActivity) > 15*time.Minute {
			delete(activeSessions, session)
		}
	}
	
	// Обновляем количество онлайн
	visitorStats.OnlineNow = len(activeSessions)
	
	// Выводим статистику в терминал
	printVisitorStats()
}

// Функция для периодической очистки старых сессий
func startSessionCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			<-ticker.C
			visitorsMutex.Lock()
			
			now := time.Now()
			cleanedCount := 0
			for session, lastActivity := range activeSessions {
				if now.Sub(lastActivity) > 15*time.Minute {
					delete(activeSessions, session)
					cleanedCount++
				}
			}
			
			// Обновляем количество онлайн после очистки
			visitorStats.OnlineNow = len(activeSessions)
			
			visitorsMutex.Unlock()
			
			if cleanedCount > 0 {
				fmt.Printf("\n🧹 Очищено %d неактивных сессий", cleanedCount)
				printVisitorStats()
			}
		}
	}()
}

// Middleware для отслеживания посещений
func trackVisitorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Отслеживаем посетителя только для HTML страниц
		if r.URL.Path == "/" || 
		   r.URL.Path == "/account" || 
		   r.URL.Path == "/tap" || 
		   r.URL.Path == "/contacts" || 
		   r.URL.Path == "/login" || 
		   r.URL.Path == "/register" || 
		   r.URL.Path == "/easycoin" {
			trackVisitor(r)
		}
		next(w, r)
	}
}

func home_page(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %s", r.URL.Path)
	
	tmpl, err := template.ParseFiles("templates/homepage.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	gameMutex.Lock()
	isLoggedIn := currentUser != ""
	var user *User
	var totalBalance float64
	
	if isLoggedIn {
		user = appData.Users[currentUser]
		// Суммируем баланс пользователя и баланс с тапалки
		totalBalance = user.Balance + appData.TapBalance
	}
	gameMutex.Unlock()

	data := struct {
		Title        string
		User         *User
		Tasks        []Task
		Stats        map[string]int
		TapBalance   float64
		TotalBalance float64
		IsLoggedIn   bool
		IsGuest      bool
	}{
		Title:        "CryptoTasks - Зарабатывай Easy Coin",
		User:         user,
		Tasks:        []Task{
			{
				ID:          1,
				Title:       "Подписка на Twitter",
				Description: "Подпишитесь на наш Twitter и сделайте ретвит",
				Reward:      "1.5 EC",
				Difficulty:  "easy",
				Category:    "social",
				Status:      "available",
			},
			{
				ID:          2,
				Title:       "Telegram Community",
				Description: "Вступите в наше Telegram сообщество",
				Reward:      "2.0 EC",
				Difficulty:  "easy",
				Category:    "social",
				Status:      "available",
			},
			{
				ID:          3,
				Title:       "Bug Bounty",
				Description: "Найдите уязвимости в нашем смарт-контракте",
				Reward:      "100 EC",
				Difficulty:  "hard",
				Category:    "development",
				Status:      "available",
			},
			{
				ID:          4,
				Title:       "Content Creation",
				Description: "Создайте видео-обзор нашей платформы",
				Reward:      "50 EC",
				Difficulty:  "medium",
				Category:    "content",
				Status:      "completed",
			},
		},
		Stats:        map[string]int{
			"total_users":   15420,
			"active_tasks":  23,
			"online_now":    visitorStats.OnlineNow, // Используем реальные данные
		},
		TapBalance:   appData.TapBalance,
		TotalBalance: totalBalance,
		IsLoggedIn:   isLoggedIn,
		IsGuest:      !isLoggedIn,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func account_page(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %s", r.URL.Path)
	
	// Проверка авторизации
	if currentUser == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	tmpl, err := template.ParseFiles("templates/account.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	gameMutex.Lock()
	user := appData.Users[currentUser]
	gameMutex.Unlock()

	data := struct {
		Title   string
		User    *User
		Stats   map[string]interface{}
		IsGuest bool
	}{
		Title: "Мой аккаунт - CryptoTasks",
		User:  user,
		Stats: map[string]interface{}{
			"total_earned":    user.TotalEarned + appData.TapBalance,
			"tasks_completed": user.Completed,
			"success_rate":    user.SuccessRate,
			"current_streak":  user.CurrentStreak,
			"referrals":       user.Referrals,
		},
		IsGuest: false,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func tap_page(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %s", r.URL.Path)
	
	// Проверка авторизации
	if currentUser == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	tmpl, err := template.ParseFiles("templates/tap.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	gameMutex.Lock()
	currentBalance := appData.TapBalance
	user := appData.Users[currentUser]
	gameMutex.Unlock()

	data := struct {
		Title   string
		Balance float64
		User    *User
	}{
		Title:   "Тапалка Easy Coin - CryptoTasks",
		Balance: currentBalance,
		User:    user,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func contacts_page(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %s", r.URL.Path)
	
	tmpl, err := template.ParseFiles("templates/contacts.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		Email string
		TG    string
	}{
		Title: "Контакты - CryptoTasks",
		Email: "support@cryptotasks.com",
		TG:    "@cryptotasks_support",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func login_page(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %s", r.URL.Path)
	
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Передаем информацию об ошибке если есть
	hasError := r.URL.Query().Get("error") == "1"
	
	data := struct {
		Title    string
		HasError bool
	}{
		Title:    "Вход - CryptoTasks",
		HasError: hasError,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func register_page(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %s", r.URL.Path)
	
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Передаем информацию об ошибке если есть
	hasError := r.URL.Query().Get("error") == "exists"
	
	data := struct {
		Title    string
		HasError bool
	}{
		Title:    "Регистрация - CryptoTasks",
		HasError: hasError,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func easycoin_page(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %s", r.URL.Path)
	
	tmpl, err := template.ParseFiles("templates/easycoin.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Easy Coin - Информация",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func login_handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		
		gameMutex.Lock()
		defer gameMutex.Unlock()
		
		// Простая проверка (в реальном приложении нужно хэширование паролей)
		if user, exists := appData.Users[username]; exists {
			// В демо-версии проверяем любой непустой пароль
			if password != "" {
				currentUser = username
				user.IsGuest = false
				saveData()
				http.Redirect(w, r, "/account", http.StatusSeeOther)
				return
			}
		}
		
		// Если пользователь не найден или неправильный пароль
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}
	
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func register_handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		// Убраны неиспользуемые переменные password и email
		r.FormValue("password") // читаем но не используем
		r.FormValue("email")    // читаем но не используем
		
		gameMutex.Lock()
		defer gameMutex.Unlock()
		
		// Проверяем, не занято ли имя пользователя
		if _, exists := appData.Users[username]; exists {
			http.Redirect(w, r, "/register?error=exists", http.StatusSeeOther)
			return
		}
		
		// Создаем нового пользователя
		appData.Users[username] = &User{
			Username:     username,
			Balance:      50.0, // Начальный бонус
			Completed:    0,
			MemberSince:  time.Now().Format("02 January 2006"),
			Level:        1,
			Rank:         "Новичок",
			TotalEarned:  50.0,
			SuccessRate:  0,
			CurrentStreak: 0,
			Referrals:    0,
			IsGuest:      false,
		}
		
		currentUser = username
		saveData()
		
		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}
	
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func logout_handler(w http.ResponseWriter, r *http.Request) {
	gameMutex.Lock()
	currentUser = ""
	gameMutex.Unlock()
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Обработчик для тапа
func tap_handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST %s", r.URL.Path)
	
	// Проверка авторизации
	if currentUser == "" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": false, "message": "Требуется авторизация"}`)
		return
	}
	
	if r.Method == "POST" {
		gameMutex.Lock()
		appData.TapBalance += 0.1 // Добавляем EC при каждом тапе
		currentBalance := appData.TapBalance
		
		// Обновляем общий заработок пользователя
		if user, exists := appData.Users[currentUser]; exists {
			user.TotalEarned += 0.1
		}
		
		gameMutex.Unlock()
		
		saveData()

		log.Printf("Tap! Новый баланс: %.2f EC", currentBalance)

		// Возвращаем обновленный баланс
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "balance": %.2f}`, currentBalance)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// Обработчик для получения баланса
func balance_handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %s", r.URL.Path)
	
	gameMutex.Lock()
	currentBalance := appData.TapBalance
	gameMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"balance": %.2f}`, currentBalance)
}

// Обработчик для сброса баланса
func reset_balance_handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST %s", r.URL.Path)
	
	if r.Method == "POST" {
		gameMutex.Lock()
		appData.TapBalance = 0
		gameMutex.Unlock()
		
		saveData()

		log.Println("Баланс сброшен!")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true}`)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// API для получения статистики
func stats_handler(w http.ResponseWriter, r *http.Request) {
	visitorsMutex.Lock()
	defer visitorsMutex.Unlock()
	
	stats := map[string]interface{}{
		"total_visitors":  visitorStats.TotalVisitors,
		"unique_visitors": visitorStats.UniqueVisitors,
		"online_now":      visitorStats.OnlineNow,
		"active_sessions": len(activeSessions),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func HandleReq() {
	// Запускаем очистку сессий
	startSessionCleaner()
	
	// Статические файлы
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	
	// Страницы с middleware для отслеживания
	http.HandleFunc("/", trackVisitorsMiddleware(home_page))
	http.HandleFunc("/account", trackVisitorsMiddleware(account_page))
	http.HandleFunc("/tap", trackVisitorsMiddleware(tap_page))
	http.HandleFunc("/contacts", trackVisitorsMiddleware(contacts_page))
	http.HandleFunc("/login", trackVisitorsMiddleware(login_page))
	http.HandleFunc("/register", trackVisitorsMiddleware(register_page))
	http.HandleFunc("/easycoin", trackVisitorsMiddleware(easycoin_page))
	
	// Обработчики форм
	http.HandleFunc("/api/login", login_handler)
	http.HandleFunc("/api/register", register_handler)
	http.HandleFunc("/api/logout", logout_handler)
	
	// API для тапалки
	http.HandleFunc("/api/tap-action", tap_handler)
	http.HandleFunc("/api/get-balance", balance_handler)
	http.HandleFunc("/api/reset-balance", reset_balance_handler)
	
	// API для статистики
	http.HandleFunc("/api/stats", stats_handler)

	fmt.Println("🚀 CryptoTasks запущен на http://localhost:8080")
	fmt.Println("👤 Страница аккаунта: http://localhost:8080/account")
	fmt.Println("🔐 Регистрация: http://localhost:8080/register")
	fmt.Println("📞 Контакты: http://localhost:8080/contacts")
	fmt.Println("🎮 Тапалка Easy Coin: http://localhost:8080/tap")
	fmt.Println("📊 API статистики: http://localhost:8080/api/stats")
	fmt.Println("\n📊 Статистика посещений будет отображаться здесь:")
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	HandleReq()
}