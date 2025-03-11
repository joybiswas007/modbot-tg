# ModBot-TG
A Telegram bot written in Go that tracks message counts, rewards points based on activity, and allows users to redeem bonuses from the shop.

**Note:** For better performance, make sure to promote the bot to an admin. And bot ONLY works in group.

## WIP
This bot is still in development—features and functionality might change over time.

## Bot Commands

```
/start - Introduces the bot and explains its purpose.
/help - Lists all available commands and their descriptions.
/id - Displays the chat ID and user ID.
/gift userid amount - gift some of your bp to other users. Example: /gift 1234566 10
or reply to a user whom you want to send the gift: /gift amount or /gift 10
/shop - display all the items available in shop
/seize userid amount - As penaly seize some points from users
/seize amount - reply to users message whose points are we going to deduct
/boost - display users available boost (User can buy only one boost at a time)
/buy itemID - buy any item specified by item id
/stats - Send `/stats` to view your own stats. Reply to another user's message with `/stats` to see their stats.
/rank type - Displays the leaderboard for a specific period. Available types: `daily`, `weekly`, `monthly`. Example: `/rank daily` to view the daily rankings.
/history - Shows the last `50` number of messages where the user earned points. (Admin ONLY)
```


## Installation

1. **Clone the repository**  
   ```
    git clone https://github.com/joybiswas007/modbot-tg.git 
    cd modbot-tg
    ```
2. **Copy the example config file**
    ```
    cp example.modbot.yaml .modbot.yaml
    ```
3. **Edit `.modbot.yaml` and configure the bot settings as needed.**

## MakeFile
Build the application
```bash
make build
```
Run the application
```bash
make run
```
Run the test suite:
```bash
make test
```
Clean up binary from the last build:
```bash
make clean
```

Runing the bot:
```
./modbot
```
or pass the config file
```
./modbot -config /path/to/modbot.yaml
```

**If you don't specify a config path, the bot will look for .modbot.yaml in the directory from which it is running.**

## Note on Database
No manual setup is required for the database. Upon starting the bot, the database will be automatically migrated.


## BP System (Bonus Points)

The bot tracks message counts and rewards users with points based on their activity in the group. Below is how the default point system works:

### **Default Point Distribution**
- **Text Messages:** Users receive a random number of points between `1` and `5`.
- **Photo:** `+1` point  
- **Document:** `+1` point  
- **Animation (GIFs):** `+1` point  
- **Audio:** `+1` point  
- **Sticker:** `+1` point  

These values are **configurable**, meaning the group owner can modify them based on their preferences.  

The more active a user is, the more points they earn, which can be used to redeem rewards from the shop.

### TODO

#### 🚀 Upcoming Features
- ~~[ ] **Buy Boost** - Users can purchase temporary boosts to earn extra points.~~
- [ ] **Gift Boost** - Users can gift boost benefits to other members.
- [ ] **Bonus Pool** - A shared pool where users contribute points, later distributed as rewards.
- [ ] **Global Double Bonus Event** - A special event where all point earnings are doubled for a limited time.

#### ⚠ Rule Enforcement
- ~~[ ] **Penalty Command** - Admins can deduct points from users who break group rules.~~


## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

Show your support by starring [⭐️](https://github.com/joybiswas007/modbot-tg/stargazers) this project!
