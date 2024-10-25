import asyncio
import logging
import os
from time import mktime, gmtime

import boto3
import feedparser

from tg.tg import TelegramBot


def update_last_published(table, url, new_last_published):
    table.update_item(
        Key={
            'url': url
        },
        UpdateExpression="set last_published = :lp",
        ExpressionAttributeValues={
            ':lp': new_last_published
        }
    )


def lambda_handler(event: any, context: any):
    dynamodb = boto3.resource('dynamodb')
    table = dynamodb.Table(os.environ["DYNAMODB_TABLE"])
    records_sent = 0

    tg = TelegramBot(os.environ["TELEGRAM_TOKEN"])

    for item in table.scan()['Items']:
        logging.info(f"Processing {item['url']}")

        try:

            # parse feed
            parsed_item = feedparser.parse(item['url'])

            # check that there is a value for last published
            if item['last_published'] == 0:
                update_last_published(table, item['url'], int(mktime(parsed_item.entries[0].published_parsed)))

                continue

            parsed_last_published = gmtime(int(item['last_published']))

            # Form an array of entries that are newer than the last published
            new_entries_stack = []
            for entry in parsed_item.entries:
                if entry.published_parsed <= parsed_last_published:
                    break

                new_entries_stack.append(entry)

            # Go over all new entries and send them to channel
            new_last_published = 0
            for entry in reversed(new_entries_stack):
                logging.info(f"Found new post for {item['url']} published at {entry.published}")

                asyncio.run(tg.send_message(os.environ["TELEGRAM_CHAT_ID"],
                                            tg.format_message(entry.title, parsed_item.feed.title, entry.summary,
                                                              entry.link)))

                records_sent += 1
                new_last_published = int(mktime(entry.published_parsed))

            # Update the last published if changed
            if new_last_published > 0:
                update_last_published(table, item['url'], new_last_published)
        except Exception as e:
            logging.error(f"Error processing {item['url']}: {e}")

    return {"status": "ok", "records_sent": records_sent}
