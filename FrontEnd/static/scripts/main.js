const blueIndicator = document.querySelector(".blue-square-indicator");
let lastTrasform = 0;

function animateNavbar(id) {
    const navbarId = document.getElementById(id);

    const navbarElement = navbarId.getBoundingClientRect();
    const indicatorElement = blueIndicator.getBoundingClientRect();
    let currentTrasform = parseInt(navbarElement.left + (navbarId.offsetWidth / 2) - (indicatorElement.left + (blueIndicator.offsetWidth / 2)));

    blueIndicator.style.transform = `translateX(${currentTrasform + lastTrasform}px)`;
    lastTrasform += currentTrasform;

    scrollDown(id);
}

function scrollDown(currentId) {
    if (document.body.offsetWidth < 481 && currentId != "profile") closeOnClick();
    
    switch(currentId) {
        case "matches":
            document.getElementById("matches-container").scrollIntoView({
                behavior: 'smooth'
            });
            break;
        case "statistics":
            document.getElementById("statistics-container").scrollIntoView({
                behavior: 'smooth'
            });
            break;
        case "teams":
            document.getElementById("teams-container").scrollIntoView({
                behavior: 'smooth'
            });
            break;
        case "profile":
            document.getElementById("burger-profile").style.backgroundPosition = " 0 100%";
            goToBot();
            if (document.body.offsetWidth > 480) blueIndicator.style.width = "160px";
            break;
    }
}

window.addEventListener('scroll', () => {
    if (window.scrollY > 120 && blueIndicator.style.transform !== "") {
        blueIndicator.style.transform = `translateX(0px)`;
        lastTrasform = 0;
    }
});

window.addEventListener('resize', () => {
    blueIndicator.style.transform = `translateX(0px)`;
    lastTrasform = 0;
});

//burger-menu
const hamButton = document.querySelector(".burger-menu-icon");
const popup = document.querySelector(".popup-menu");

hamButton.addEventListener("click", hambHandler);

function hambHandler(e) {
  popup.classList.add("popup-open");
  document.body.classList.toggle("noscroll");
}

function closeOnClick() {
  popup.classList.remove("popup-open");
  document.body.classList.remove("noscroll");
}

function openRules() {
    let overlay = document.getElementById("overlay");
    let popup = document.getElementById("rules-popup");

    overlay.classList.remove("hiden-rules");
    popup.classList.remove("hiden-rules");
    popup.style.top = `50%`;
    document.body.classList.toggle("noscroll");

    overlay.addEventListener("click", () => {
        popup.style.top = `45%`;
        overlay.classList.add("hiden-rules");
        popup.classList.add("hiden-rules");
        document.body.classList.remove("noscroll");
    });

    document.addEventListener('keydown', event => {
        if (event.key === 'Escape') {
            popup.style.top = `45%`;
            overlay.classList.add("hiden-rules");
            popup.classList.add("hiden-rules");
        }
    });
}

function closeRules() {
    let overlay = document.getElementById("overlay");
    let popup = document.getElementById("rules-popup");

    popup.style.top = `45%`;
    overlay.classList.add("hiden-rules");
    popup.classList.add("hiden-rules");
    document.body.classList.remove("noscroll");
}

const STATISTICS_API = "http://77.239.124.241:8080/statistics";

// Функция для загрузки и отображения статистики
async function fetchStatistics() {
    try {
        const response = await fetch(STATISTICS_API);

        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }

        const statistics = await response.json();
        statistics.sort((a, b) => b.points - a.points)

        updateStatisticsContainer(statistics);
    } catch (error) {
        console.error("Ошибка загрузки статистики:", error);
    }
}

// Функция обновления блока статистики
function updateStatisticsContainer(statistics) {
    const statisticsGrid = document.querySelector(".table-grid-rows");

    const playOffElement = document.createElement("div");
    playOffElement.classList.add("play-off-line");
    playOffElement.innerHTML = `
        <div class="title">Плей-офф</div>
        <div class="separation-line"></div>
        <img src="../assets/icons/Header-arrow.svg" alt="arrow">`;

    statisticsGrid.innerHTML = "";

    statistics.forEach((team, index) => {
        if (index === 4) {
            statisticsGrid.appendChild(playOffElement);
            return;
        }

        const teamRow = document.createElement("div");
        teamRow.className = "table-names table-cell";

        switch(index) {
            case 0:
                teamRow.style.backgroundColor = "#3F6FFF";
                break;
            case 1:
                teamRow.style.backgroundColor = "#2B2B2B";
                break;
            case 2:
                teamRow.style.backgroundColor = "#1B1B1B";
                break;
            default:
                teamRow.style.border= "1px solid #343434";
                break;
        }

        teamRow.innerHTML = `
            <div class="team-info">
                <div class="match-team-name" style='margin-right: 1vw;'>
                    ${getShortName(team.name)}
                </div>
                <div>${minimizeTeamName(team.name)}</div>
                <div class="hide-team-name">${getShortName(team.name)}</div>
            </div>
            <div class="centered-points">
                <div class="name-of-point">${team.games}</div>
                <div class="name-of-point">${team.wins}</div>
                <div class="name-of-point">${team.losses}</div>
                <div class="name-of-point"><span ${index === 0 ? 'style="color: #FFF;"' : 'style="color: #3F6FFF;"'}>${team.points}</span></div>
            </div>
        `;

        statisticsGrid.appendChild(teamRow);
    });
}

// Загружаем данные при загрузке страницы
document.addEventListener("DOMContentLoaded", () => {
    fetchMatches();
    fetchStatistics();
});