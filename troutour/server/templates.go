package server

// Don't edit this file. It's automatically generated!

var tmplS = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Troubador Tour Board</title>
  <link rel="stylesheet" href="https://ajax.googleapis.com/ajax/libs/jquerymobile/1.4.5/jquery.mobile.css">
  <link rel="stylesheet" href="/client/s.css">
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.js"></script>
  <script src="https://ajax.googleapis.com/ajax/libs/jquerymobile/1.4.5/jquery.mobile.js"></script>
  <script>
    var userID = '{{.UserID}}';
    var googleAuthURL = '{{.GoogleAuthURL}}';
  </script>
  <script src="client/s.js"></script>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
  <canvas id="glcanvas" style="z-index: -1; width: 100%;">
    Your browser doesn't appear to support the <code>&lt;canvas&gt;</code> element.
  </canvas>
  <div id="activity-zone" class="ui-overlay-shadow">
    <button id="ping" class="ui-btn filter-auth">Ping</button>
    <a id="login_google" class="ui-btn filter-nauth">Log in with Google</a>
  </div>
  <div id="gazetteer">
    ☛ Tap the map to show info about tapped region here
  </div>
  <div id="messages">
  </div>

  <a id="show-junkdrawer" class="filter-auth ui-btn ui-icon-bars ui-btn-icon-left" data-rel="popup" href="#junkdrawer-menu">Misc</a>
  <div id="junkdrawer-menu" data-role="popup">
  <ul data-role="listview" data-inset="true">
    <li><a href="/logout" class="ui-btn ui-icon-refresh ui-btn-icon-left">Log out</a>
    <li><a id="show-inventory" href="#" class="ui-btn ui-icon-user ui-btn-icon-left">Show</a>
  </ul>
  </div>

  <div id="inventory" data-role="popup">
    <div data-role="header"><h2>Inventory</h2><a href="#" data-rel="back" data-role="button" data-theme="a" data-icon="delete" data-iconpos="notext" class="ui-btn-right">Close</a></div>
    <div role="main" class="ui-content">
    <table>
      <tr><td> 💎 <td> <span id="inv-prisms"></span>
      <tr><td> 💰 <td> <span id="inv-coins"></span>
      <tr><td> 🏆 <td> <span id="inv-trophies"></span>
      <tr><td> Cr <td> <span id="inv-cred"></span>
    </table>
    </div>
  </div>

  <div id="welcome" data-role="popup">
    <div data-role="header"><h2>Troubador Tour Board</h2><a href="#" data-rel="back" data-role="button" data-theme="a" data-icon="delete" data-iconpos="notext" class="ui-btn-right">Close</a></div>
    Welcome to the game. How it goes:
    <ul>
      <li>You are a theatrical agent for musicians.
	You want <i>🎤&nbsp;Clients</i>.
      <li>Walk around. Press the button to <i>visit</i> the nearby region.
      <li>Visit a region to get a <i>💎Prism</i> for that region.
      <li>Visit a region near another region for which you have a 💎Prism
	to establish a <i>⛗Route</i> between those regions. This "spends"
	the 💎Prism.
      <li>You want <i>🏆Trophies</i>.
      <li>When you visit a region, you might recruit a 🎤&nbsp;Client.
	If so, they travel with you for a while and then settle down
	when you visit an empty region.
      <li>When you visit a region, your 🎤&nbsp;Clients pay you 💰.
	⛗Route-connected 🎤&nbsp;Clients pay more 💰 than unconnected
	🎤&nbsp;Clients do.
      <li>If you have many <i>🏆Trophies</i> and lots of 💰, you can spend
	them on a party to which you invite all the right people, thus
	boosting your professional credibility.
      <li>Sometimes regions appear where there were none.
	Sometimes regions disappear. This can tear up your ⛗Routes.
    </ul>
    <p><b>Privacy Policy:</b> <br>
      <i>Data we collect:</i>
      This game "remembers" some places you've
      been close to&mdash;places you've been in the last hour, places
      you have 💎Prisms for, places you have ⛗Routes, places you have
      🎤&nbsp;Clients. <br>
      <i>How we use it:</i> To make the game rules work, as described above. <br>
      <i>Who do we share it with:</i> Law enforcement, if they have a warrant, hypothetically. <br>
    </div>
</body>
</html>

`
