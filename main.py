from twscrape import API

api = API()

async def main():
    async for tweet in api.search("$SOL lang:en", limit=50):
        print(tweet.rawContent)

import asyncio
asyncio.run(main())