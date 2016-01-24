# BitX

## Bitcoin market making bot
This trading bot executes trades based on a predefined strategy. To run it, register an account at https://bitx.co/ and supply an API key and secret as command line args.

**TODO**
- remove human-interaction requirement (for safety)
- handle rate limiting from the API
- track executed orders to determine profit/loss
- create more bot strategies

## BitX streamer
Use gRPC to stream changes to the BitX market in real-time. Consists of a client and server component.