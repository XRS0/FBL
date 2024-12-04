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
            window.location.href = "https://vk.com/artist/0pokhval";
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

// URL для получения данных о матчах и статистике
const MATCHES_API = "http://localhost:8080/matches";
const STATISTICS_API = "http://localhost:8080/statistics";

// Функция для загрузки и отображения матчей
async function fetchMatches() {
    try {
        const response = await fetch(MATCHES_API);

        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }

        const matches = await response.json();
        updateMatchesContainer(matches);
    } catch (error) {
        console.error("Ошибка загрузки матчей:", error);
    }
}

// Функция для загрузки и отображения статистики
async function fetchStatistics() {
    try {
        const response = await fetch(STATISTICS_API);

        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }

        const statistics = await response.json();
        updateStatisticsContainer(statistics);
    } catch (error) {
        console.error("Ошибка загрузки статистики:", error);
    }
}

// Функция обновления блока матчей
function updateMatchesContainer(matches) {
    const matchesGrid = document.querySelector(".matches-grid-block");

    // Очищаем существующий контент
    matchesGrid.innerHTML = "";

    matches.forEach(match => {
        const matchWidget = document.createElement("div");
        matchWidget.className = "match-widget";

        matchWidget.innerHTML = `
            <img src="../assets/images/thumbnail.png" alt="thumb">
            <div class="match-info">
                <div class="match-time">${match.time}</div>
                <div class="match-score">${match.team1_score}:${match.team2_score}</div>
                <div class="match-status">
                    <span style="text-decoration: underline;">${match.status}</span>
                </div>
            </div>
            <img src="../assets/images/thumbnail.png" alt="thumb">
        `;

        matchesGrid.appendChild(matchWidget);
    });
}

// Функция обновления блока статистики
function updateStatisticsContainer(statistics) {
    const statisticsGrid = document.querySelector(".table-grid-rows");

    // Очищаем существующий контент
    statisticsGrid.innerHTML = "";

    statistics.forEach((team, index) => {
        const teamRow = document.createElement("div");
        teamRow.className = "table-names table-cell";
        teamRow.style.backgroundColor = index === 0 ? "#3F6FFF" : "#2B2B2B";

        teamRow.innerHTML = `
            <div class="team-info">
                <img src="../assets/images/thumbnail.png" alt="team-icon" class="team-logo-margin">
                <div>${team.name}</div>
                <div class="hide-team-name">${team.abbreviation}</div>
            </div>
            <div class="centered-points">
                <div class="name-of-point">${team.games}</div>
                <div class="name-of-point">${team.wins}</div>
                <div class="name-of-point">${team.losses}</div>
                <div class="name-of-point"><span style="color: #3F6FFF;">${team.points}</span></div>
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
