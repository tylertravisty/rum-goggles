# Doing

Next steps:
- delete page needs to handle new architecture
- activatePage: verify defer page.activeMu.Unlock does not conflict with display function

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
