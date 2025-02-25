package suggestion

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SuggestionService struct {
	UsersCollection *mongo.Collection
	Ctx             context.Context
}

func NewSuggestionService(usersCollection *mongo.Collection, ctx context.Context) *SuggestionService {
	return &SuggestionService{
		UsersCollection: usersCollection,
		Ctx:             ctx,
	}
}

func (s *SuggestionService) HandleSuggestionCommand(session *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		session.ChannelMessageSend(message.ChannelID, "❌ Usage: !suggestion <role/company>")
		return
	}

	query := strings.Join(args, " ")
	suggestions, err := s.FetchSuggestions(query)
	if err != nil {
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("❌ Error fetching suggestions: %v", err))
		return
	}

	if len(suggestions) == 0 {
		session.ChannelMessageSend(message.ChannelID, "❌ No matching users found.")
		return
	}

	var response strings.Builder
	response.WriteString("🔍 Suggestions:\n")
	for _, suggestion := range suggestions {
		response.WriteString(fmt.Sprintf("- **%s** | **%s**\n", suggestion["name"], suggestion["company"]))
	}

	session.ChannelMessageSend(message.ChannelID, response.String())
}

func (s *SuggestionService) FetchSuggestions(query string) ([]map[string]interface{}, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"role": bson.M{"$regex": query, "$options": "i"}},
			{"company": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := s.UsersCollection.Find(s.Ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(s.Ctx)

	var users []map[string]interface{}
	if err := cursor.All(s.Ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}