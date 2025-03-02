const TEAMS_API = "";   // link to /teams

function updateTeamsContainer(teams) {
  const teamsContainer = document.getElementById("teams-container");
  const teamColors = [...TeamColors].concat(shuffle(TeamColors.slice(teams.length - TeamColors.length)));  // беру стандартные 6 цветов и дополняю их еще парой рандомных из 6 цветов (работает максимум на 12 команд)
  let playersPictures = [...PlayersPictures].concat(PlayersPictures.slice(teams.length - PlayersPictures.length));   // здесь создаю экземпляр картинок игроков

  teamsContainer.innerHTML = "";

  teams.forEach((team, teamIndex) => {
    const teamBlock = document.createElement("div");
    teamBlock.classList.add("team-widget");
    teamBlock.style.float = teamIndex % 2 == 0 ? "left" : "right";

    shuffle(playersPictures);

    teamBlock.innerHTML = `
      <div class="team-info team-header">
        <img src=${team.logo}>
        <div class="team-name"><span style="color: ${teamColors[teamIndex]};">${team.name.length < 15 ? team.name : minimizeTeamName(team.name)}</span> Team</div>
      </div>

      <div class="separation-line" style="border: 1px solid ${teamColors[teamIndex]}; box-shadow: 0px 2px 6px rgba(0, 0, 0, 0.9);"></div>

      <div class="players-grid">
        ${team.players.map((player, playerIndex) =>
          `<div class="player-card" onMouseOver="this.style.borderColor='${teamColors[teamIndex]}'" onMouseOut="this.style.borderColor='#343434'">
            <img src=${playersPictures[playerIndex]} alt="player">
            <div class="player-number">${player.number}</div>
            <div class="separation-line"></div>
            <div class="player-name">${player.name.length < 12 ? player.name : minimizeLastName(player.name)}</div>
          </div>
        `).join("")}
      </div>`;

    teamsContainer.appendChild(teamBlock);
  });
}