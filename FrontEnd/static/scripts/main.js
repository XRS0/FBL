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
// 89.104.69.138
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
        matches.sort((a, b) => new Date(a.time) - new Date(b.time));

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
        statistics.sort((a, b) => b.points - a.points)

        updateStatisticsContainer(statistics);
    } catch (error) {
        console.error("Ошибка загрузки статистики:", error);
    }
}

let createdConditions = [];

// Функция обновления блока матчей
function updateMatchesContainer(matches) {
    const matchesGrid = document.getElementById("matches-container");

    // Очищаем существующий контент
    matchesGrid.innerHTML = "";

    matches.forEach(match => {
        const matchWidget = document.createElement("div");
        matchWidget.className = "match-widget";

        matchWidget.innerHTML = `
            <div class="match-team-name">
                ${getShortName(match.team1_name)}
            </div>
            <div class="match-info">
                <div class="match-time">${formatDateTime(match.time).time}</div>
                <div class="match-score">${match.team1_score}:${match.team2_score}</div>
                <div class="match-status">
                    <span style="text-decoration: underline;">${match.status}</span>
                </div>
            </div>
            <div class="match-team-name">
                ${getShortName(match.team2_name)}
            </div>
        `;

        let isHeaderExist = createdConditions.some(i => i.time.slice(0, 9) == match.time.slice(0, 9) && i.loc == match.loc);
        
        if (!isHeaderExist) {
            createdConditions.push({ time: match.time, loc: match.loc });
            matchesGrid.appendChild(wrapMatches(matchWidget.innerHTML, match.time, match.loc));
        } else {
            let matchesGrid = document.querySelectorAll(".matches-grid-block");
            matchesGrid = matchesGrid[matchesGrid.length - 1];
            matchesGrid.appendChild(matchWidget);
        }
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
    `;
    return assembledContainer;
}

function getShortName(name) {
    if (name.split(" ").length == 1) return name.slice(0, 2);
    let [fWord, sWord] = name.split(" ");
    return (fWord[0] + sWord[0]).toUpperCase();
}

function formatDateTime(timestamp) {
    // Создаём объект даты из строки
    const dateObj = new Date(timestamp);

    // Массив названий месяцев
    const months = [
        "января", "февраля", "марта", "апреля", "мая", "июня",
        "июля", "августа", "сентября", "октября", "ноября", "декабря"
    ];

    // Проверка на корректность даты
    if (isNaN(dateObj.getTime())) {
        return { date: "Некорректная дата", time: "Некорректное время" };
    }

    // Извлекаем день, месяц и время
    const day = dateObj.getDate();
    const month = months[dateObj.getMonth()]; // Название месяца
    const hours = String(dateObj.getHours()).padStart(2, '0'); // Часы (добавляем 0 при необходимости)
    const minutes = String(dateObj.getMinutes()).padStart(2, '0'); // Минуты (добавляем 0 при необходимости)

    // Форматируем результаты
    const formattedDate = `${day} ${month}`;
    const formattedTime = `${hours}:${minutes}`;

    return { date: formattedDate, time: formattedTime };
}

function goToBot() {
    var url = "https://t.me/BFBLB_bot"; 
    var windowName = "_blank"; 
    var windowFeatures = "width=800,height=600"; 

    window.open(url, windowName, windowFeatures);
}

function minimizeTeamName(name) {
    if (name.length > 16) {
        let minimized = name.split(" ")
        return minimized.map(i => i[0]).join("");
    } else return name;
}


// Загружаем данные при загрузке страницы
document.addEventListener("DOMContentLoaded", () => {
    fetchMatches();
    fetchStatistics();
});

document.querySelectorAll(".social-tg").forEach(function (element) {
    element.addEventListener("click", function () {
        const url = "https://t.me/BFBLB_bot"; 
        window.location.href = url;
    });
});