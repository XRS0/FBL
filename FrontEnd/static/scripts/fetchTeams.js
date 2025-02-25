async function fetchTeams() {
  try {
      const response = await fetch("http://77.239.124.241:8080/statistics", {
      method: "GET",
      mode: "cors"
  });
      if (!response.ok) {
          throw new Error(`Ошибка HTTP: ${response.status}`);
      }

      const statistics = await response.json();
      statistics.sort((a, b) => b.points - a.points);
      
      TEAMS = [...TEAMS, ...statistics]

      if (document.title === "Fast Break League") updateStatisticsContainer(TEAMS);
      
      if (document.title === "Fast Break League") updateTeamsContainer(TEAMS.slice(0, 2));
      else updateTeamsContainer(TEAMS);
  } catch (error) {
      console.error("Can't response data from TEAMS-API:", error);
  }
}