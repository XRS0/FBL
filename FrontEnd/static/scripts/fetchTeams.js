async function fetchTeams() {
  try {
      const response = await fetch(TEAMS_API, {
      method: "GET",
      mode: "cors"
  });
      if (!response.ok) {
          throw new Error(`Ошибка HTTP: ${response.status}`);
      }

      const statistics = await response.json();
      statistics.sort((a, b) => b.points - a.points);
      
      TEAMS = [...TEAMS, ...statistics]
      fetchMatches();

      if (document.title === "Fast Break League") updateStatisticsContainer(TEAMS);
      
      if (document.title === "Fast Break League") updateTeamsContainer(TEAMS);
      else updateTeamsContainer(TEAMS);
  } catch (error) {
      console.error("Can't response data from TEAMS-API:", error);
  }
}
