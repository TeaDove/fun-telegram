package tgdtos

type DTO interface {
	TGType() TGType
}

type Record[T DTO] struct {
	Type    TGType `json:"type"`
	Payload T      `json:"payload"`
}

func NewRecord[T DTO](payload T) Record[T] {
	return Record[T]{Type: payload.TGType(), Payload: payload}
}

type HistoryHeader struct {
	Version int    `json:"version"`
	PeerID  int64  `json:"peer_id,string"`
	Title   string `json:"title"`
}

func (HistoryHeader) TGType() TGType {
	return TGTypeHistoryHeader
}

type Author struct {
	ID       int64  `json:"id,string"`
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
}

func (Author) TGType() TGType {
	return TGTypeAuthor
}

type Topic struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func (Topic) TGType() TGType {
	return TGTypeTopic
}

type Message struct {
	ID       int    `json:"id"`
	Date     int64  `json:"date"`
	AuthorID int64  `json:"author_id,string,omitempty"`
	Text     string `json:"text,omitempty"`
	ReplyTo  int    `json:"reply_to,omitempty"`
	TopicID  int    `json:"topic_id,omitempty"`
}

func (Message) TGType() TGType {
	return TGTypeMessage
}
