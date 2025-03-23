const isDev = window.location.hostname === "localhost";
const wsPath = "ws://" + window.location.hostname + "/detective-club/ws"
const socket = new WebSocket(wsPath);

const activePlayerBtn = document.getElementById("active-player-btn");
const inputBoxArea = document.getElementById("input-box-area");
const submitBtn = document.getElementById("submit-btn");
const closeBtn = document.getElementById("close-btn");
let userId = "0";

// Show the input box and disable the button
activePlayerBtn.addEventListener("click", () => {
    inputBoxArea.style.display = "block";
    message = {
        purpose: "claim-active-player",
        content: userId
    };
    socket.send(JSON.stringify(message));
});

submitBtn.addEventListener("click", () => {
    // Get the input value
    const inputValue = document.getElementById("input-box").value;

    message = {
        purpose: "give-clue",
        content: inputValue
    }
    // Send the input value to the server
    socket.send(JSON.stringify(message));

    // Clear the input box
    document.getElementById("input-box").value = "";
});

// Close the input box and re-enable the button
closeBtn.addEventListener("click", () => {
    inputBoxArea.style.display = "none";
    message = {
        purpose: "unclaim-active-player",
        content: userId
    };
    socket.send(JSON.stringify(message));
});

socket.addEventListener("message", (event) => {
    const data = event.data;
    message = JSON.parse(data);

    switch (message.purpose) {
        case "set-id":
            userId = message.content;
            break;
        case "set-active-player":
            if (message.content == "") {
                activePlayerBtn.disabled = false; // Enable the button for all players
            } else {
                activePlayerBtn.disabled = true; // Disable the button for all players
            }
            break;
        case "give-clue":
            alert(message.content);
            break;
        case "inform-num-players":
            document.getElementById("num-players").innerText = `Number of players: ${message.content}`;
            break;
        default:
            break;
    }
});