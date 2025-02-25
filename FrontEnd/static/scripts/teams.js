const TEAMS_API = "http://localhost:8080/teams_data";

function updateTeamsContainer(teams) {
  const teamsContainer = document.getElementById("teams-container");
  const teamColors = [...TeamColors].concat(shuffle(TeamColors.slice(teams.length - TeamColors.length)));  // беру стандартные 6 цветов и дополняю их еще парой рандомных из 6 цветов (работает максимум на 12 команд)
  let playersPictures = [...PlayersPictures];   // здесь создаю экземпляр картинок игроков

  teamsContainer.innerHTML = "";

  teams.forEach((team, teamIndex) => {
    const teamBlock = document.createElement("div");
    teamBlock.classList.add("teams-grid-block");

    shuffle(playersPictures);

    teamBlock.innerHTML = `
      <div class="team-widget" onMouseOver="this.style.borderColor='${teamColors[teamIndex]}'" onMouseOut="this.style.borderColor='#343434'">
        <div class="team-info team-header">
          <div class='logo-container'>
            <img src=${team.logo}>
          </div>
          <div class="team-name"><span style="color: ${teamColors[teamIndex]};">${team.name}</span> Team</div>
        </div>

        <div class="separation-line" style="border: 1px solid ${teamColors[teamIndex]}; box-shadow: 0px 2px 6px rgba(0, 0, 0, 0.9);"></div>

        <div class="players-grid">
          ${team.players.map((player, playerIndex) => 
            `<div class="player-card" onMouseOver="this.style.borderColor='${teamColors[teamIndex]}'" onMouseOut="this.style.borderColor='#343434'">
              <img src=${playersPictures[playerIndex]} alt="player">
              <div class="player-number">${player.number}</div>
              <div class="separation-line"></div>
              <div class="player-name">${player.name}</div>
            </div>
          `).join("")}
        </div>
      </div>
    `;

    teamsContainer.appendChild(teamBlock);
  });
}
