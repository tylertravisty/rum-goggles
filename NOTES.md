# Roadmap

Rum Goggles App:

v1-beta:
- Chat polls
- Stream statistics

v1.N:
- Download subscriber data, track re-subs, trigger rules on re-subs
- Stream moderator bot
- OBS integration

Rum Goggles Service:
- Monitor all live stream chats
- Channel points
- Spam/Troll tracking

# Bugs

Updating chat bot URL does not show up in UI until app is restarted

Sign in with email does not work

Chat bot rule menu back button does not work in macOS

If connection to chat stream breaks, gracefully handle error
- try to reconnect
- let the chat rules know, etc.
- test with VPN

# Doing

Add bypass to commands for:
- Host, admin, mod, etc.

Monitor how many handlers are listening to a producer.
- If producer.Stop is called, subtract from count.
- If count == 0, stop producer

Change API producer to monitor changes and only send new events, one at a time, to app, instead of the entire response; create datatype for single API event
- update chatbot and page handlers to use single API events
- page details should add to activity list one at a time
    - store page list in Go, send entire list to frontend on updates
    - list can be updated by any producer

Add timeouts to event triggers to prevent rate limit?

Don't stop rule if chat error is 429 Too Many Requests

Check if sender is logged in before running rule. If not, return rule error.

Add max rant amount for commands

Button to export log file -> user selects folder

Style scroll bars on Windows
- WebView2 issue

Indicator in chatbot that producer is running
    - this can be in many different places as needed

Custom stream moderator rules
- block on keywords
- block on regex/pattern
- blocks can be: timed, stream, forever

Next steps:
- enable/disable rules from starting, including start all/stop all buttons
- delete page needs to handle new architecture
    - app.producers.Active(*name) instead of app.producers.ApiP.Active(*name)
    - app.producers.Stop(*name)
- activatePage: verify defer page.activeMu.Unlock does not conflict with display function
- in chatbot list, show indicator if any rules in chatbot are running

For Dashboard page,
- Api or chat producer could error, need to be able to start/restart both handlers on error
- Show user error if api or chat stop/error

On API errors
- include backoff multiple, if exceeded then stop API

Add option to delete API key for accounts?

Add better styles/icon to account details menu

Start screen:
- check for new updates and tell user

Trigger on event from API vs. trigger on event from chat
- Chat bot trigger on follow requires API key
    - Give user warning when setting up trigger on follow that it will only work with accounts/channels for which user has saved an API key

Reset session information in config on logout

Show error when choosing file "chooseFile"
Show filename in chat bot list
Add styling to choose file button

Commands
- specify for follower/subscriber/locals only/rants
    - check badges for subscriber and locals

Update
- github.com/rhysd/go-github-selfupdate
- github.com/inconshreveable/go-update

When API key is added, loading indicator freezes all user interactions. Need to give user a graceful way to stop/close add channel if it breaks.

If api query returns error:
- stop interval
- show error to user
- wait for user to press "retry" button to restart interval

Settings
- allow user to change api key
- allow user to change api interval time

Get user's: username, password, stream key
Query API
Display followers, subscribers, etc.

User settings:
- API query timer (default: 2s)

# To Do

Currently relies on Rumble to manage account username case-sensitivity.
- Change database table to use UNIQUE COLLATE NOCASE on account username