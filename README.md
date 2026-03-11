# Skitgubbe
Skitgubbe, also called Palace, is a swedish card game where the objective is to get rid of all your cards. The first player to do so wins. You start off with 3 cards in hand and 6 cards on the table, 3 cards face down and 3 cards face up on top of them. You place cards onto a pile and draw cards everytime you place some. When the draw pile is empty, you begin to use the cards on the table. When all cards are gone, you win. 

I hosted the site on [skitgubbe.dev.fly](https://skitgubbe.fly.dev/) using [fly.io](fly.io).

## Motivation
I created this project so me and my friends can play it together. When I started this project, we couldn't find a similar game so I took it upon myself and made one. Now, we have something to play between other games.

## Features
- Create an account
- Host a game of skitgubbe and invite your friends using a generated code
- Play the game

## Quick Start
1. Navigate to [skitgubbe.dev.fly](https://skitgubbe.fly.dev/) and create an account.
2. Invite your friends ant start playing! 

## Usage
If you want to host it yourself and make it better, this section is for you.

Environment variables to set: 
- "DB_PATH" - path to the database file

To build and start the server:
``` go build -o out && ./out ```

The server is setup at port 8080, but you can changeit to whatever you want. Navigate to:
[localhost:8080](localhost:8080)

## Contributing
