## About

This simple telegram bot polls atom subscription of the specified project or
a group and sends all activity updates to a telegram chat id.

## Usage

```sh
go get -u github.com/ilyaglow/gitlab-atom-tgbot
```

You should get telegram bot token from `@BotFather` first. Then you should
[find out the group chat ID](https://stackoverflow.com/a/32572159) and the link
to atom activity subscription - go to https://your-gitlab-repo/activity and
copy a link similar to this:
`https://gitlab.com/your-user/your-repo.atom?feed_token=TOKEN`.

Use environment variables to configure the bot:
```sh
	TGBOT_TOKEN=<telegram_bot_token> \
	TG_CHAT_ID=<telegram_chat_id> \
	GITLAB_ATOM_LINK=<link_to_gitlab_activity_atom> \
	./gitlab-atom-tgbot
```
