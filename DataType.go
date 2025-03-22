package drive

type User struct {
	Id          string   `bson:"id" json:"id"`
	Name        string   `bson:"name" json:"name"`
	Role        string   `bson:"role" json:"role"`
	Permissions []string `bson:"permissions" json:"permissions"`
	Password    string   `bson:"password" json:"password"`
}

type Resource struct {
	Id        string   `bson:"id" json:"id"`
	Name      string   `bson:"name" json:"name"`
	OwnerId   string   `bson:"ownerId" json:"ownerId"`
	SharedId  []string `bson:"sharedId" json:"sharedId"`
	Location  string   `bson:"location" json:"location"`
	Type      string   `bson:"type" json:"type"`
	Content   []string `bson:"content" json:"content"`
}