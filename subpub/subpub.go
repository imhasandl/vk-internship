package subpub

import "context"

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cd MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

// func () Subscribe(subject string, cd MessageHandler) {
// 	return 
// }


func NewSubPub() SubPub {
	return nil
}