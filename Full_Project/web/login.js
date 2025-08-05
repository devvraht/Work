document.addEventListener("DOMContentLoaded", () => {
  // Hardcoded credentials and div count
  const hardcodedUser = "admin";
  const hardcodedPass = "1234";
  const divCount = 6; // Number of divs to create on dashboard

  document.getElementById("login").addEventListener("click", () => {
    // Simulate login with hardcoded credentials
    const username = hardcodedUser;
    const password = hardcodedPass;

    if (username === "admin" && password === "1234") {
      sessionStorage.setItem("loggedIn", "true");
      sessionStorage.setItem("divCount", divCount);
      window.location.href = "dashboard.html";
    } else {
      alert("Login failed");
    }
  });
});
