import json
import os
from time import mktime, gmtime

import boto3
import feedparser


def lambda_handler(event: any, context: any):
    dynamodb = boto3.resource('dynamodb')
    table = dynamodb.Table(os.environ["DYNAMODB_TABLE"])
    records_sent = 0

    for item in table.scan()['Items']:
        parsed_item = feedparser.parse(item['url'])
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
            # TODO: Send to channel

            records_sent += 1
            new_last_published = mktime(entry.published_parsed)

        # Update the last published if changed
        if new_last_published > 0:
            table.update_item(
                Key={
                    'url': item['url']
                },
                UpdateExpression="set last_published = :lp",
                ExpressionAttributeValues={
                    ':lp': new_last_published
                }
            )

    return {"status": "ok", "records_sent": records_sent}
