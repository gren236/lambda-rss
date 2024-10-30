import html

from telegram import Bot
from telegram.constants import ParseMode


class TelegramBot:
    template = """
<b>{title}</b>
<i>From {author}</i>

{content}

<a href="{link}">{link}</a>
    """

    token: str

    def __init__(self, token: str):
        self.token = token

    async def send_message(self, chat_id: str, message: str):
        await Bot(token=self.token).send_message(chat_id=chat_id, text=message, parse_mode=ParseMode.HTML)

    def format_message(self, title: str, author: str, content: str, link: str):
        content = html.escape(content)

        if len(content) > 1000:
            content = content[:1000 - 3] + '...'

        return self.template.format(title=title, author=author, content=content, link=link)
