# Roadmap

Rum Goggles App:
- Chat bot rule triggers on stream events
    - On: follow, subscribe, rant, raid
- Chat polls
- Stream statistics
- Stream moderator bot

Rum Goggles Service:
- Channel points
- Spam/Troll tracking

# Doing

Before v1-alpha release:
- stop all running rules when chat bot is deleted
- indicator in chatbot that producer is running
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
