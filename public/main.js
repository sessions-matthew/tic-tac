const boardId = "board2";

const host = "localhost:8080";

function createBoard(id, game, player) {
  var url = `ws://${host}/listen/${game}/${player}`;
  var c = new WebSocket(url);

  var open = false;
  var buttons = [];

  function setPlayer(c) {
    player = c;
  }

  function markWinner(done, winner, player) {
    if (done) {
      if (winner == player) $(id).attr("class", "winner");
      else $(id).attr("class", "loser");
    }
  }

  function getBoard(id, gameId) {
    fetch(`/game/${gameId}`, {
      method: "GET",
    }).then((res) => {
      res.json().then((game) => {
        const board = game.board;
        console.log(board);

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
            btn.onclick = () => sendMove(x, y);

            rowButtons.push(btn);
            r.appendChild(btn);
          });

          buttons.push(rowButtons);

          $(id).append(r);
        });
        markWinner(game.isDone, game.winner, player);
      });
    });
  }

  function sendMove(x, y) {
    if (open) {
      const t = player;
      c.send(JSON.stringify({ x, y, t }));
    }
  }

  c.onmessage = function (msg) {
    console.log(msg.data);

    var json = JSON.parse(msg.data);

    switch (json.type) {
      case "player":
        console.log("got player event");

        break;
      case "move":
        const { x, y, t } = json.content;
        buttons[y][x].textContent = t;
        break;
      case "game":
        switch (json.event) {
          case "over":
            markWinner(true, json.content, player);
            break;
        }
        break;
    }
  };

  c.onopen = function () {
    open = true;
  };

  getBoard(id, game);
}

const playerToken = $("#game")[0].attributes.token.value;
const username = $("#game")[0].attributes.username.value;
const gameid = $("#game")[0].attributes.gameid.value;
createBoard("#board0", gameid, playerToken);
