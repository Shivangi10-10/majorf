package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Title string `json:"title"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GraphResponse struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func GraphHandler(users *mongo.Collection, connections *mongo.Collection, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var usersList []bson.M
		var connectionsList []bson.M

		cursor, _ := users.Find(ctx, bson.M{})
		cursor.All(ctx, &usersList)

		cursor2, _ := connections.Find(ctx, bson.M{})
		cursor2.All(ctx, &connectionsList)

		nodes := make([]Node, 0)
		for _, user := range usersList {
			name := user["name"].(string)
			company := user["company"].(string)
			role := user["role"].(string)
			tooltip := fmt.Sprintf("Company: %s\nRole: %s", company, role)
			nodes = append(nodes, Node{
				ID: name, 
				Label: name,
				Title: tooltip,
			})
		}

		edges := make([]Edge, 0)
		for _, conn := range connectionsList {
			edges = append(edges, Edge{
				From: conn["user1"].(string),
				To:   conn["user2"].(string),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GraphResponse{Nodes: nodes, Edges: edges})
	}
}
