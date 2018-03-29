package login

import (
	"log"
	"net/http"
	"os"
	"errors"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"github.com/garyburd/go-oauth/oauth"
	"encoding/json"
	"net/url"
)

const(
	TwRequestTokenKey = "TwitterRequestToken"
	TwRequestSecretKey = "TwitterRequestToken"
	TwOauthTokenKey = "TwitterOauthToken"
	TwOauthSecretKey = "TwitterOauthToken"
	TwitterApiEndpoint = "https://api.twitter.com/1.1/"
)

var(
	/*
		Twitterの認証方式は以下3方式
    1.ユーザだけの認証 (Implicit grant)
    2.アプリケーションだけの認証 (Client Credentials Grant)
    3.ユーザとアプリケーション両方の認証 (Authorization Code Grant)
		2以外はoauth2非対応
	 */
	TwitterConfig = struct {
		RedirectURL string
		Endpoint map[string]interface{}
		Credential oauth.Credentials
	}{
		RedirectURL: os.Getenv("TWITTER_REDIRECT_URL"),
		Endpoint: map[string]interface{}{
			"RequestURI": "https://api.twitter.com/oauth/request_token",
			"AuthorizationURI": "https://api.twitter.com/oauth/authorize",
			"TokenRequestURI": "https://api.twitter.com/oauth/access_token",
		},
		Credential: oauth.Credentials{
			Token: os.Getenv("TWITTER_CONSUMER_KEY"),
			Secret: os.Getenv("TWITTER_CONSUMER_SECRET"),
		},
	}
)

type Twitter struct {
	w	http.ResponseWriter
	r *http.Request
	config *oauth.Client
	tok *oauth.Credentials
	client *http.Client
}

func NewTwitter(w http.ResponseWriter, r *http.Request) *Twitter {
	config := &oauth.Client {
		TemporaryCredentialRequestURI: TwitterConfig.Endpoint["RequestURI"].(string),
		ResourceOwnerAuthorizationURI: TwitterConfig.Endpoint["AuthorizationURI"].(string),
		TokenRequestURI: TwitterConfig.Endpoint["TokenRequestURI"].(string),
		Credentials: TwitterConfig.Credential,
	}
	client := urlfetch.Client(appengine.NewContext(r))
	twitter := &Twitter{w, r, config, nil, client}

	sess := NewSession(w, r, &sessionConfig)
	requestToken := sess.Get(TwOauthTokenKey)
	requestSecret := sess.Get(TwOauthSecretKey)

	if requestToken != "" && requestSecret != "" {
		twitter.tok = &oauth.Credentials{
			Token: requestToken,
			Secret: requestSecret,
		}
	}
	return twitter
}


func (this *Twitter) Login() {
	sess := NewSession(this.w, this.r, &sessionConfig)
	log.Printf("sess ID %s", sess.session.ID)
	// TODO Set state token For CSRF attack check
	tok, err := this.config.RequestTemporaryCredentials(this.client, TwitterConfig.RedirectURL, nil)
	if err != nil {
		panic(err)
	}
	log.Printf("RequestToken %#v", tok)
	sess.Set(TwRequestTokenKey, tok.Token)
	sess.Set(TwRequestSecretKey, tok.Secret)
	sess.Save()

	url := this.config.AuthorizationURL(tok, nil)
	log.Printf("URL %s", url)
	http.Redirect(this.w, this.r, url, http.StatusFound)
}

func (this *Twitter) Callback(redirect string) {
	sess := NewSession(this.w, this.r, &sessionConfig)

	oauthToken := this.r.FormValue("oauth_token") // TODO verify
	log.Printf("token:%s", oauthToken)
	verifier := this.r.FormValue("oauth_verifier")
	log.Printf("verifier:%s", verifier)
	requestToken := sess.Get(TwRequestTokenKey)
	requestSecret := sess.Get(TwRequestSecretKey)
	if requestToken == "" || requestSecret == "" || verifier == "" {
		log.Printf("Twitter Callback Token error requestToken:%s requestSecret:%s verifir:%s", requestToken, requestSecret, verifier)
		http.Redirect(this.w, this.r, "/login/twitter", http.StatusFound)
		return
	}

	tok, _, err := this.config.RequestToken(this.client, &oauth.Credentials{
		Token: oauthToken,
		Secret: requestSecret,
	}, verifier)
	if err != nil {
		log.Printf("Twitter RequestToken error maybe requestToken expired")
		http.Redirect(this.w, this.r, "/login/twitter", http.StatusFound)
		return
	}
	log.Printf("Twitter Token %#v", tok)

	this.tok = tok
	sess.Set(TwOauthTokenKey, tok.Token)
	sess.Set(TwOauthSecretKey, tok.Secret)
	sess.Save()

	// Tget Me
	tw, err := this.getMe()
	if err != nil {
		log.Printf("People Get me error:%#v", err)
		panic(err)
	}
	log.Printf("People Get me %#v", tw)
	sess.Set("id", tw["id_str"].(string))
	var email string = ""
	if tw["email"] != nil && tw["email"] != "" {
		// アプリでの許可とログイン時の許可があった場合だけ取得可能
		email = tw["email"].(string)
	}
	sess.Set("email", email)
	sess.Set("name", tw["screen_name"].(string))
	sess.Set("photo", tw["profile_image_url_https"].(string))
	sess.Save()
	http.Redirect(this.w, this.r, redirect, http.StatusFound)
}

func (this *Twitter) getMe() (map[string]interface{}, error) {
	if this.config == nil {
		return nil, errors.New("client not set")
	}
	if this.tok == nil {
		return nil, errors.New("token not set")
	}
	log.Printf("Tok %#v", this.tok)

	v := url.Values{}
	v.Set("include_email", "true") // Change app permission https://apps.twitter.com/ to request user's email
	resp, err := this.config.Get(this.client, this.tok, TwitterApiEndpoint+"account/verify_credentials.json", v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("Response:%s", resp.Body)

	if resp.StatusCode >= 500 {
		return nil, errors.New("Twitter is unavailable")
	}

	if resp.StatusCode >= 400 {
		return nil, errors.New("Twitter request is invalid")
	}

	var result map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	log.Printf("Result %#v", result)
	return result, nil
}
