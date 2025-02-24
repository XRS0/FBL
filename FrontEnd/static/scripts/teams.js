const TEAMS_API = "http://xxx:8080/teams";

async function fetchTeams() {
  try {
      const response = await fetch(TEAMS_API);

      if (!response.ok) {
        throw new Error(`Ошибка HTTP: ${response.status}`);
      }

      const teams = await response.json();
 
      updateTeamsContainer(teams);
  } catch (error) {
      console.error("Ошибка загрузки статистики:", error);
  }
}

function updateTeamsContainer(teams) {
  const teamsContainer = document.getElementById("teams-container");

  teamsContainer.innerHTML = "";

  teams.forEach(team => {
    const teamBlock = document.createElement("div");
    teamBlock.className = "teams-block";

    teamBlock.innerHTML = `
      <div class="teams-grid-block">
        <div class="team-widget">
          <div class="team-info team-header">
            <img src=${team.logo}>
            <div class="team-name"><span style="color: #FF1F62;">${team.name}</span> Team</div>
          </div>

          <div class="separation-line" style="border: 1px solid #FF1F62; box-shadow: 0px 2px 6px rgba(0, 0, 0, 0.9);"></div>

          <div class="players-grid">
            ${team.players.map(player => {
              const playerCard = document.createElement("div");
              playerCard.classList.add("player-card");

              playerCard.innerHTML = `
                <img src=${player.avatar} alt="player">
                <div class="separation-line"></div>
                <div class="player-name">${player.name}</div>
              `;
            })}
          </div>
        </div>
      </div>
    `;

    teamsContainer.appendChild(teamBlock);
  });
}