const boardId = "board2";

const host = "localhost:8080";

function createBoard(id, game, player) {
  var url = `ws://${host}/game/${game}/${player}`;
  var c = new WebSocket(url);

  var gameView = {
		buttons: [],
		drawBoard: (controller, id, game) => {
			const board = game.board;
			controller.players = game.players;

			board.forEach((row, y) => {
				const r = document.createElement("div");
				var rowButtons = [];
				row.forEach((col, x) => {
					const btn = document.createElement("div");
					btn.textContent = col;
					btn.style.display = "inline-block";
					btn.style.verticalAlign = "top";
					btn.style.width = "50px";
					btn.style.height = "50px";
					btn.style.margin = "2px";
					btn.style.backgroundColor = "grey";
					btn.onclick = () => controller.sendMove(x, y);

					rowButtons.push(btn);
					r.appendChild(btn);
				});

				gameView.buttons.push(rowButtons);
				$(id).append(r);
			});
		},
    setTile: (x, y, t) => {
			gameView.buttons[y][x].textContent = t;
    },
    setStatus: (status) => {
      $("#status").text(status);
    },
    setWinner: (winner, player) => {      
			if (winner == player) $(id).attr("class", "winner");
			else $(id).attr("class", "loser");
    },
	};
  
	function fetchBoard(gameId) {
		return fetch(`/game/${gameId}`, {
			method: "GET",
		});
	}

	var gameController = {
		players: [],
		addPlayer: (player) => {
			gameController.players.push(player);
		},
		disconnectPlayer: (player) => {
      gameView.setStatus(`${player} has disconnected`);
		},
		sendMove: (x, y) => {
			const t = player;
			c.send(JSON.stringify({ x, y, t }));
		},
		evalStatus: () => {
			if (gameController.players.length < 2) {
				gameView.setStatus("Waiting for player...");
			} else {
				gameView.setStatus("");
			}
		},
    evalWinner: (done, winner, player) => {
      if(done) {
        gameView.setWinner(winner, player);
        gameView.setStatus(`${winner} has won`);
      }
    },
		getBoard: (id, gameId) => {
			fetchBoard(gameId).then((res) => {
				res.json().then((game) => {
          console.log(game);
					gameView.drawBoard(gameController, id, game);
					gameController.evalStatus();
					gameController.evalWinner(game.isDone, game.winner, player);
				});
			});
		},
		processEvent: (msg) => {
			console.log(msg.data);
			var json = JSON.parse(msg.data);

			switch (json.type) {
				case "player":
					switch (json.event) {
						case "connected":
							gameController.addPlayer(json.content);
							gameController.evalStatus(gameController);
							break;
						case "disconnect":
							gameController.disconnectPlayer(json.content);
							break;
					}
					break;
				case "move":
					const { x, y, t } = json;
					gameView.setTile(x, y, t);
					break;
				case "movenext":
					const { username } = json;
					gameView.setStatus(`It is ${username}'s turn`);
					break;
				case "game":
					switch (json.event) {
						case "over":
							gameController.evalWinner(true, json.winner, player);
							break;
					}
					break;
			}
		}
	};

	c.onmessage = gameController.processEvent;

  c.onopen = function () {
    gameController.getBoard(id, game);
  };
}

const playerToken = $("#game")[0].attributes.token.value;
const username = $("#game")[0].attributes.username.value;
const gameid = $("#game")[0].attributes.gameid.value;
createBoard("#board0", gameid, playerToken);
