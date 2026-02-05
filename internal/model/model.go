package model

type MessageResponse struct {
	ID           int64  `json:"id"`
	Author       string `json:"author"`
	Content      string `json:"content"`
	IsOwnerLiked bool   `json:"is_owner_liked"`
	IsLiked      bool   `json:"is_liked"`
	IsDisliked   bool   `json:"is_disliked"`
	Likes        int    `json:"likes"`
	Dislikes     int    `json:"dislikes"`
}

type SvgMessageModel struct {
	ID           int64
	Author       string
	Content      string
	Likes        int
	Dislikes     int
	IsOwnerLiked bool
}
