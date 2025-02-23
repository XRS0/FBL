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

function getShortName(name) {
  if (name.split(" ").length == 1) return name.slice(0, 2);
  let [fWord, sWord] = name.split(" ");
  return (fWord[0] + sWord[0]).toUpperCase();
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

document.querySelectorAll(".social-tg").forEach(function(element) {

  element.addEventListener("click", function() {
      const url = element.classList.contains("social-vk") ? "https://vk.com/fastbreakleague" : "https://t.me/BFBLB_bot";
      window.location.href = url;
  });
});

function shuffle(array) {
  let currentIndex = array.length;

  // While there remain elements to shuffle...
  while (currentIndex != 0) {

    // Pick a remaining element...
    let randomIndex = Math.floor(Math.random() * currentIndex);
    currentIndex--;

    // And swap it with the current element.
    [array[currentIndex], array[randomIndex]] = [
      array[randomIndex], array[currentIndex]];
  }
}