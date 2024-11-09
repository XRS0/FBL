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