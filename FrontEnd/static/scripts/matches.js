const MATCHES_API = "http://77.239.124.241:8080/matches";

// Функция для загрузки и отображения матчей
async function fetchMatches() {
  try {
      const response = await fetch(MATCHES_API);

      if (!response.ok) {
          throw new Error(`Ошибка HTTP: ${response.status}`);
      }

      const matches = await response.json();
      matches.sort((a, b) => new Date(b.time) - new Date(a.time));

      if (document.title === "Fast Break League") updateMatchesContainer(matches.slice(0, 6));
      else updateMatchesContainer(matches);
  } catch (error) {
      console.error("Ошибка загрузки матчей:", error);
  }
}

let createdConditions = [];

// Функция обновления блока матчей
function updateMatchesContainer(matches) {
    const matchesGrid = document.getElementById("matches-container");

    matchesGrid.innerHTML = "";

    matches.forEach(match => {
        const matchWidget = document.createElement("div");
        matchWidget.className = "match-widget";

        matchWidget.innerHTML = `
            <div class="match-team-name">
                ${TEAMS.find(team => team.name === match.team1_name && team.logo != undefined)
                    ? `<img src=${TEAMS.find(team => team.name === match.team1_name).logo}>`
                    : getShortName(match.team1_name)
                }
            </div>
            <div class="match-info">
                <div class="match-time">${formatDateTime(match.time).time}</div>
                <div class="match-score">${match.team1_score}:${match.team2_score}</div>
                <div class="match-status">
                    <span">${match.status}</span>
                </div>
            </div>
            <div class="match-team-name">
                ${TEAMS.find(team => team.name === match.team1_name && team.logo != undefined)
                    ? `<img src=${TEAMS.find(team => team.name === match.team2_name).logo}>`
                    : getShortName(match.team2_name)
                }
            </div>
        `;
        
        let isHeaderExist = createdConditions.some(i => i.time.slice(0, 10) === match.time.slice(0, 10) && i.loc == match.loc);
        
        if (!isHeaderExist) {
            createdConditions.push({ time: match.time, loc: match.loc });
            matchesGrid.appendChild(wrapMatches(matchWidget.innerHTML, match.time, match.loc));
        } else {
            let matchesGrid = document.querySelectorAll(".matches-grid-block");
            matchesGrid = matchesGrid[matchesGrid.length - 1];
            matchesGrid.prepend(matchWidget);
        }
    });
}

function wrapMatches(container, date, location) {
  let assembledContainer = document.createElement("div");
  assembledContainer.classList.add("container");
  assembledContainer.innerHTML = `
    <div class="matches-info-bar">
        <div>
            <h2 id="matches-date">${formatDateTime(date).date}</h2>
            <h3 id="matches-location">${location}</h3>
        </div>
            
        <div class="matches-hint">
            Время проведения <br>
            Счет игры <br>
            Статус
        </div>
    </div>
    <div class="separation-line"></div>

    <div class="matches-block">
        <div class="matches-grid-block">
            <div class="match-widget">
                ${container}
            </div>
        </div>
    </div>

    ${document.title === "Fast Break League" ?
      `<div class="teams-block" style="margin-bottom: 50px;">
          <div class="fillness-button" onclick="window.open('./matches.html', '_self')">Больше матчей</div>
      </div>` : ""
    }
  `;
  return assembledContainer;
}
