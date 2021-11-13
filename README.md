# vulnerable-click-game
An online click game written in golang that deliberately contains vulnerabilities for learning purposes.

# Vulnerabilities

> Please stop reading here, if you don't want to spoil your fun by knowing the vulnerabilities. If you want some pointers, just keep on reading.

The following vulnerabilities have intentionally be implemented within this application:
1. Broken Access Control (A01:2021) – Access to all files in the repository is possible via the webserver. It is possible to show a different users profile by user ID enumeration.
2. Injection (A03:2021) – Using a SQL Injection, Login without known password is possible. A XSS attack is possible when submitting points for a game.
5. Insecure Design (A04:2021) – Submitting arbitrary points for a game is possible as the endpoint directly takes the "game result" as it calculated from the frontend.
3. Security Misconfiguration (A05:2021) – No security headers are configured whatsoever which increases the attack surface e.g. for XSS attacks.
4. Vulnerable and Outdated Components (A06:2021) – An outdated version of JQuery is used that contains known vulnerabilities (might not be able to be exploited as the vulnerable methods are not in use here).
