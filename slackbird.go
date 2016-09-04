package slackbird

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/Rompei/inco"
)

const (
	Tweet    = "tweet"
	Follow   = "follow"
	Unfollow = "unfollow"
	Retweet  = "retweet"
	Favorite = "favorite"
	Delete   = "delete"
	DM       = "dm"
)

// SlackBird is object for handling commands.
type SlackBird struct {
	api        *anaconda.TwitterApi
	webhookURL string
}

// NewSlackBird is constructor.
func NewSlackBird(consumerKey, consumerSecret, accessToken, accessTokenSecret, webhookURL string) *SlackBird {
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessTokenSecret)
	return &SlackBird{
		api:        api,
		webhookURL: webhookURL,
	}
}

// Do do the subcommand.
func (sb *SlackBird) Do(text, channel string, errCh chan error) (err error) {
	t := strings.SplitN(strings.TrimSpace(text), " ", 2)
	if len(t) < 1 {
		err = errors.New("Sub command doesn't exist")
		sb.sendErrorMessage(err.Error(), channel)
		if errCh != nil {
			errCh <- err
		}
		return err
	}
	switch t[0] {
	case Tweet:
		err = sb.tweet(t)
	case Follow:
		err = sb.follow(t)
	case Unfollow:
		err = sb.unfollow(t)
	case Retweet:
		err = sb.retweet(t)
	case Favorite:
		err = sb.favorite(t)
	case Delete:
		err = sb.del(t)
	case DM:
		err = sb.dm(t)
	default:
		err = fmt.Errorf("Unknown command %s", t[0])
	}
	if err != nil {
		sb.sendErrorMessage(err.Error(), channel)
	}
	if errCh != nil {
		errCh <- err
	}
	return
}

func (sb *SlackBird) sendErrorMessage(text, channel string) {
	msg := &inco.Message{
		Text:    text,
		Channel: channel,
	}
	inco.Incoming(sb.webhookURL, msg)
}

func (sb *SlackBird) tweet(t []string) error {
	if len(t) < 2 {
		return errors.New("Argument is not enough")
	}
	tweetText := t[1]
	_, err := sb.api.PostTweet(tweetText, url.Values{})
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("Failed to tweet %s", tweetText)
	}
	return nil
}

func (sb *SlackBird) follow(t []string) error {
	if len(t) < 2 {
		return errors.New("Argument is not enough")
	}
	userName := t[1]
	_, err := sb.api.FollowUser(userName)
	if err != nil {
		return fmt.Errorf("Could not find user %s", userName)
	}
	return nil
}

func (sb *SlackBird) unfollow(t []string) error {
	if len(t) < 2 {
		return errors.New("Argument is not enough")
	}
	userName := t[1]
	_, err := sb.api.UnfollowUser(userName)
	if err != nil {
		return fmt.Errorf("Could not find user %s", userName)
	}
	return nil
}

func (sb *SlackBird) retweet(t []string) error {
	if len(t) < 2 {
		return errors.New("Argument is not enough")
	}
	id, err := sb.getIDFromURL(t[1])
	if err != nil {
		return fmt.Errorf("User id must be integer")
	}
	if _, err = sb.api.Retweet(id, true); err != nil {
		return fmt.Errorf("Could not retweet tweet %s", t[1])
	}
	return nil
}

func (sb *SlackBird) favorite(t []string) error {
	if len(t) < 2 {
		return errors.New("Argument is not enough")
	}
	id, err := sb.getIDFromURL(t[1])
	if err != nil {
		return fmt.Errorf("User id must be integer")
	}
	if _, err = sb.api.Favorite(id); err != nil {
		return fmt.Errorf("Could not favorite tweet %s", t[1])
	}
	return nil
}

func (sb *SlackBird) del(t []string) error {
	if len(t) < 2 {
		return errors.New("Argument is not enough")
	}
	id, err := sb.getIDFromURL(t[1])
	if err != nil {
		return fmt.Errorf("User id must be integer")
	}
	if _, err = sb.api.DeleteTweet(id, true); err != nil {
		return fmt.Errorf("Could not favorite tweet %s", t[1])
	}
	return nil
}

func (sb *SlackBird) dm(t []string) error {
	if len(t) < 2 {
		return errors.New("Argument is not enough")
	}
	tt := strings.SplitN(t[1], " ", 2)
	if len(tt) < 2 {
		return errors.New("Argument is not enough")
	}
	userName := tt[0]
	text := tt[1]
	if _, err := sb.api.PostDMToScreenName(text, userName); err != nil {
		return fmt.Errorf("Could not send DM to %s %s", userName, text)
	}
	return nil
}

func (sb *SlackBird) getIDFromURL(u string) (int64, error) {
	us := strings.Split(u, "/")
	_id := us[len(us)-1]
	return strconv.ParseInt(_id, 10, 64)
}
