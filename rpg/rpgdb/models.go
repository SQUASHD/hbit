// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package rpgdb

type Character struct {
	UserID         string `json:"user_id"`
	ClassID        string `json:"class_id"`
	CharacterLevel int64  `json:"character_level"`
	Experience     int64  `json:"experience"`
	Health         int64  `json:"health"`
	Mana           int64  `json:"mana"`
	Strength       int64  `json:"strength"`
	Dexterity      int64  `json:"dexterity"`
	Intelligence   int64  `json:"intelligence"`
}

type CharacterClass struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Item struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ItemType    string  `json:"item_type"`
	Attributes  *string `json:"attributes"`
}

type Quest struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Requirements *string `json:"requirements"`
	Rewards      *string `json:"rewards"`
}

type QuestItem struct {
	QuestID          string `json:"quest_id"`
	ItemID           string `json:"item_id"`
	QuantityRequired int64  `json:"quantity_required"`
}

type UserQuest struct {
	UserID    string `json:"user_id"`
	QuestID   string `json:"quest_id"`
	Completed bool   `json:"completed"`
}