/*!
    * Start Bootstrap - SB Admin v6.0.2 (https://startbootstrap.com/template/sb-admin)
    * Copyright 2013-2020 Start Bootstrap
    * Licensed under MIT (https://github.com/StartBootstrap/startbootstrap-sb-admin/blob/master/LICENSE)
    */
(function ($) {
    "use strict";

    // Add active state to sidbar nav links
    var path = window.location.href; // because the 'href' property of the DOM element is the absolute path
    $("#layoutSidenav_nav .sb-sidenav a.nav-link").each(function () {
        if (this.href === path) {
            $(this).addClass("active");
        }
    });

    // Toggle the side navigation
    $("#sidebarToggle").on("click", function (e) {
        e.preventDefault();
        $("body").toggleClass("sb-sidenav-toggled");
    });
})(jQuery);

/**
 * Click Game Functionality
 */
let counter = 0;

function addCounter(num) {
    counter += num;
    setDisplay();
}

function setDisplay() {
    $('#counter').text(counter);
}

function sendScore() {
    var victoryshout = prompt("Shout something about your victory:", "Hell Yeah!");
    data = {
        points: counter,
        victoryshout: victoryshout,
    };
    console.log("Sending Score: ", data);
    $.ajax({
        method: "POST",
        contentType: "application/json; charset=utf-8",
        url: "api/score",
        data: JSON.stringify(data),
      })
        .done(function( msg ) {
            alert("Your score has been added. You can directly start a new game.")
            counter = 0;
            setDisplay();
        })
        .fail(function( msg ) {
            alert("There was a problem adding your score. Please try again.")
        });
}

function getHighscore() {
    $('#dataTable').DataTable( {
        "ajax": "api/highscore",
        "columns": [
            { "data": "email" },
            { "data": "points" },
            { "data": "victoryshout"}
        ],
        "order": [[ 1, "desc" ]]
    } );
}

function getUserScores() {
    let searchParams = new URLSearchParams(window.location.search)
    let userid = searchParams.get("userid")
    $('#dataTable').DataTable( {
        "ajax": {
            "url": "api/userscores",
            "data": {
                "userid": userid
            }
        },
        "columns": [
            { "data": "email" },
            { "data": "points" },
            { "data": "victoryshout"}
        ],
        "order": [[ 1, "desc" ]]
    } );
}
