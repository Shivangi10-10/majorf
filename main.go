package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/yourusername/discord-referral-bot/suggestion"
	"github.com/yourusername/discord-referral-bot/ai_service"
	"net/http"
	"github.com/yourusername/discord-referral-bot/api"
)

var (
	mongoClient           *mongo.Client
	usersCollection       *mongo.Collection
	connectionsCollection *mongo.Collection
	ratingsCollection     *mongo.Collection
	ctx                   context.Context
)

func init() {
	// Initialize context
	ctx = context.Background()
}

func main() {
	// Load environment variables

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Connect to MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	// Ping MongoDB to verify connection
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}
	fmt.Println("✅ MongoDB Connection Successful")

	// Initialize collections
	db := mongoClient.Database("referral_db")
	usersCollection = db.Collection("users")
	connectionsCollection = db.Collection("connections")
	ratingsCollection = db.Collection("ratings")

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.Handle("/api/graph", api.GraphHandler(usersCollection, connectionsCollection, ctx))


	// Create Discord bot with proper intents
	botToken := os.Getenv("BOT_TOKEN")

	// Use all intents to ensure we have proper permissions
	intents := discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuilds

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Set intents
	dg.Identify.Intents = intents

	// Register handlers
	dg.AddHandler(messageCreate)
	dg.AddHandler(ready)

	// Open connection to Discord
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %v", err)
	}
	defer dg.Close()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	go func() {
		fmt.Println("🌐 Starting Web Server at http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	
	// Keep the bot running until CTRL-C is pressed
	select {}
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("Bot is ready! Logged in as %s\n", s.State.User.Username)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Log all received messages for debugging
	fmt.Printf("Received message: '%s' from '%s' in channel '%s'\n",
		m.Content, m.Author.Username, m.ChannelID)

	// Check if the message starts with '!'
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	// Split the message into command and arguments
	parts := strings.Fields(m.Content)
	if len(parts) == 0 {
		return
	}

	// Extract command and arguments
	command := strings.TrimPrefix(parts[0], "!")
	args := parts[1:]

	fmt.Printf("Command received: '%s' with args: %v\n", command, args)

	suggestionService := suggestion.NewSuggestionService(usersCollection, ctx)


	// Handle commands
	switch command {
	case "register":
		handleRegister(s, m, args)
	case "connect":
		handleConnect(s, m, args)
	case "find_referrer":
		handleFindReferrer(s, m, args)
	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong! Bot is working.")
	case "suggestion":
		suggestionService.HandleSuggestionCommand(s, m, args)
	case "rate_referrer":
		handleRateReferrer(s, m, args)
	case "refer_message":
		handleReferMessage(s, m, args)	
	case "help":
		helpMessage := "Available commands:\n" +
			"!register <role> <company> - Register your info\n" +
			"!connect <user1> <user2> - Connect two users\n" +
			"!find_referrer <company> - Find referrers\n" +
			"!ping - Check bot status\n" +
			"!suggestion <your text> - Send a suggestion\n" +
			"!help - Show this help message\n" +
			"!rate_referrer - Leave feedback rating\n"
		s.ChannelMessageSend(m.ChannelID, helpMessage)
	default:
		s.ChannelMessageSend(m.ChannelID, "Unknown command. Type !help for available commands.")
	}
}	

func handleRegister(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "❌ Usage: !register <role> <company>")
		return
	}

	role := args[0]
	company := args[1]
	username := m.Author.Username

	err := registerUser(username, role, company)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error registering user: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ %s registered with role: %s and company: %s", username, role, company))
}

func handleConnect(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "❌ Usage: !connect <user1> <user2>")
		return
	}

	user1 := args[0]
	user2 := args[1]

	err := connectUsers(user1, user2)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error connecting users: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔗 %s is now connected to %s", user1, user2))
}

// func handleFindReferrer(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
// 	if len(args) < 1 {
// 		s.ChannelMessageSend(m.ChannelID, "❌ Usage: !find_referrer <company>")
// 		return
// 	}

// 	company := args[0]
// 	username := m.Author.Username

// 	userDetails, err := getUserDetails(username)
// 	if err != nil {
// 		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error getting user details: %v", err))
// 		return
// 	}

// 	if userDetails == nil {
// 		s.ChannelMessageSend(m.ChannelID, "❌ You need to register first! Use !register <role> <company>")
// 		return
// 	}

// 	referrer, err := findBestReferrer(username, company)
// 	if err != nil {
// 		s.ChannelMessageSend(m.ChannelID, "❌ No referrer found. Try expanding your network with !connect.")
// 		return
// 	}

// 	if referrer == "" {
// 		s.ChannelMessageSend(m.ChannelID, "❌ No referrer found. Try expanding your network with !connect.")
// 		return
// 	}

// 	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ %s can refer you for a job at %s!", referrer, company))
// }

// func handleFindReferrer(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
// 	if len(args) < 1 {
// 		s.ChannelMessageSend(m.ChannelID, "❌ Usage: !find_referrer <company>")
// 		return
// 	}

// 	company := args[0]
// 	username := m.Author.Username

// 	referrer, err := findBestReferrer(username, company)
// 	if err != nil {
// 		s.ChannelMessageSend(m.ChannelID, "❌ No referrer found. Try expanding your network with !connect.")
// 		return
// 	}

// 	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ %s can refer you for a job at %s!", referrer, company))
// }

// HITS Algorithm for Finding the Best Referrer
// func findBestReferrer(username, targetCompany string) (string, error) {
// 	connections, err := getConnections()
// 	if err != nil {
// 		return "", err
// 	}

// 	graph := make(map[string][]string)
// 	for _, conn := range connections {
// 		user1 := conn["user1"].(string)
// 		user2 := conn["user2"].(string)
// 		graph[user1] = append(graph[user1], user2)
// 		graph[user2] = append(graph[user2], user1)
// 	}

// 	hubScores := make(map[string]float64)
// 	authScores := make(map[string]float64)

// 	for user := range graph {
// 		hubScores[user] = 1.0
// 		authScores[user] = 1.0
// 	}

// 	iterations := 10
// 	for i := 0; i < iterations; i++ {
// 		newAuthScores := make(map[string]float64)
// 		newHubScores := make(map[string]float64)

// 		for user := range graph {
// 			newAuthScores[user] = 0.0
// 			for _, neighbor := range graph[user] {
// 				newAuthScores[user] += hubScores[neighbor]
// 			}
// 		}

// 		for user := range graph {
// 			newHubScores[user] = 0.0
// 			for _, neighbor := range graph[user] {
// 				newHubScores[user] += newAuthScores[neighbor]
// 			}
// 		}

// 		hubScores = newHubScores
// 		authScores = newAuthScores
// 	}

// 	var bestReferrer string
// 	maxAuthScore := -1.0

// 	for user, score := range authScores {
// 		userDetails, _ := getUserDetails(user)
// 		if userDetails != nil && userDetails["company"].(string) == targetCompany {
// 			if score > maxAuthScore {
// 				maxAuthScore = score
// 				bestReferrer = user
// 			}
// 		}
// 	}

// 	if bestReferrer == "" {
// 		return "", fmt.Errorf("no referrer found")
// 	}

// 	return bestReferrer, nil
// }

// Database functions
func registerUser(username, role, company string) error {
	// Format role if it's a comma-separated string
	var formattedRole string
	if strings.Contains(role, ",") {
		formattedRole = role
	} else {
		formattedRole = role
	}

	// Update user or insert if not exists
	filter := bson.M{"name": username}
	update := bson.M{
		"$set": bson.M{
			"name":    username,
			"role":    formattedRole,
			"company": company,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := usersCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func connectUsers(user1, user2 string) error {
	// Create connection
	filter := bson.M{"user1": user1, "user2": user2}
	update := bson.M{
		"$set": bson.M{
			"user1": user1,
			"user2": user2,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := connectionsCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func getUserDetails(username string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := usersCollection.FindOne(ctx, bson.M{"name": username}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // User not found
		}
		return nil, err
	}

	// Remove _id field
	delete(result, "_id")
	return result, nil
}

func getConnections() ([]map[string]interface{}, error) {
	cursor, err := connectionsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var connections []map[string]interface{}
	err = cursor.All(ctx, &connections)
	if err != nil {
		return nil, err
	}

	// Remove _id field from each connection
	for i := range connections {
		delete(connections[i], "_id")
	}

	return connections, nil
}

func getAllUsers() ([]map[string]interface{}, error) {
	cursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []map[string]interface{}
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}

	// Remove _id field from each user
	for i := range users {
		delete(users[i], "_id")
	}

	return users, nil
}

// Simplified find referrer function directly in main.go
// func findBestReferrer(username, targetCompany string) (string, error) {
// 	// Get all connections
// 	connections, err := getConnections()
// 	if err != nil {
// 		return "", err
// 	}

// 	// Build a simple adjacency list for the graph
// 	graph := make(map[string][]string)
// 	for _, conn := range connections {
// 		user1 := conn["user1"].(string)
// 		user2 := conn["user2"].(string)

// 		// Add both directions since it's an undirected graph
// 		graph[user1] = append(graph[user1], user2)
// 		graph[user2] = append(graph[user2], user1)
// 	}

// 	// BFS to find a referrer
// 	visited := make(map[string]bool)
// 	queue := []string{username}
// 	visited[username] = true

// 	for len(queue) > 0 {
// 		currentUser := queue[0]
// 		queue = queue[1:]

// 		// Skip self
// 		if currentUser != username {
// 			userDetails, err := getUserDetails(currentUser)
// 			if err != nil {
// 				return "", err
// 			}

// 			// Check if this user is at the target company
// 			if userDetails != nil && userDetails["company"].(string) == targetCompany {
// 				return currentUser, nil // Found a referrer!
// 			}
// 		}

// 		// Add neighbors to the queue
// 		for _, neighbor := range graph[currentUser] {
// 			if !visited[neighbor] {
// 				visited[neighbor] = true
// 				queue = append(queue, neighbor)
// 			}
// 		}
// 	}

// 	return "", fmt.Errorf("no referrer found")
// }



//////////newwwwwwwwwwwwwwww

var rolePriority = map[string]int{
	"Manager": 5,
	"SDE3": 4,
	"SDE2": 3,
	"SDE1": 2,
	"Others": 1,
}

func handleRateReferrer(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "❌ Usage: !rate_referrer <username> <score>")
		return
	}

	username := args[0]
	score, err := strconv.Atoi(args[1])
	if err != nil || score < 1 || score > 5 {
		s.ChannelMessageSend(m.ChannelID, "❌ Score must be a number between 1 and 5.")
		return
	}

	var user bson.M
	err = usersCollection.FindOne(ctx, bson.M{"name": username}).Decode(&user)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ User not found.")
		return
	}

	_, err = ratingsCollection.InsertOne(ctx, bson.M{"user_id": user["_id"], "score": score})
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Failed to rate user.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Successfully rated %s!", username))
}

func handleFindReferrer(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		s.ChannelMessageSend(m.ChannelID, "❌ Usage: !find_referrer <company>")
		return
	}

	company := args[0]
	username := m.Author.Username

	referrer, err := findBestReferrer(username, company)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ No referrer found.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ %s is the best referrer for %s!", referrer, company))
}

func findBestReferrer(username, targetCompany string) (string, error) {
	connections, err := getConnections()
	if err != nil {
		return "", err
	}

	graph := make(map[string][]string)
	for _, conn := range connections {
		user1 := conn["user1"].(string)
		user2 := conn["user2"].(string)
		graph[user1] = append(graph[user1], user2)
		graph[user2] = append(graph[user2], user1)
	}

	hubScores := make(map[string]float64)
	authScores := make(map[string]float64)
	for user := range graph {
		hubScores[user] = 1.0
		authScores[user] = 1.0
	}

	for i := 0; i < 10; i++ {
		newAuthScores := make(map[string]float64)
		newHubScores := make(map[string]float64)
		for user := range graph {
			newAuthScores[user] = 0.0
			for _, neighbor := range graph[user] {
				newAuthScores[user] += hubScores[neighbor]
			}
		}
		for user := range graph {
			newHubScores[user] = 0.0
			for _, neighbor := range graph[user] {
				newHubScores[user] += newAuthScores[neighbor]
			}
		}
		hubScores = newHubScores
		authScores = newAuthScores
	}

	var bestReferrer string
	maxScore := -1.0
	for user, score := range authScores {
		userDetails, _ := getUserDetails(user)
		if userDetails != nil && userDetails["company"].(string) == targetCompany {
			if score > maxScore || (score == maxScore && rolePriority[userDetails["role"].(string)] > rolePriority[bestReferrer]) {
				maxScore = score
				bestReferrer = user
			}
		}
	}

	if bestReferrer == "" {
		return "", fmt.Errorf("no referrer found")
	}

	return bestReferrer, nil
}

func handleReferMessage(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
    if len(args) < 3 {
        s.ChannelMessageSend(m.ChannelID, "❌ Usage: !refer_message <referrer> <role> <company>")
        return
    }

    requester := m.Author.Username
    referrer := args[0]
    role := args[1]
    company := strings.Join(args[2:], " ") // Supports multi-word company names

    message, err := ai_service.GenerateReferralMessage(requester, referrer, role, company)
    if err != nil {
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error generating referral message: %v", err))
        return
    }

    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✉️ Here is your referral message:\n%s", message))
}
