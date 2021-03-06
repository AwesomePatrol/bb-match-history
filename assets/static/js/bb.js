const diff2str = [
  "I'm Too Young to Die",
  "Piece of Cake",
  "Easy",
  "Normal",
  "Hard",
  "Nightmare",
  "Ultra-Violence",
  "Fun and Fast"
  ];

const urlParams = new URLSearchParams(window.location.search);

function showDifficultyBreakdown() {
    let df = $("#diffBreak");
    df.empty();
    let tbl = $("<table>")
        .addClass("table")
        .addClass("table-hover")
        .addClass("table-striped");
    let th = $("<tr>");
    $.each(diff2str, function(i, val) {
        th.append($("<td>").append(val));
    });
    tbl.append($("<thead>")
        .append(th)
        .addClass("table-secondary")
    );
    let tr = $("<tr>");
    $.each(diff2str, function(i, val) {
        tr.append($("<td>").append($("#matches td:contains(" + val + ")").length));
    });
    tbl.append($("<tbody>").append(tr));
    df.append(tbl);
    df.show();
}

function fillShortMatchPlayerNameRow(team, i) {
    let td = $("<td>");
    if (i >= team.length) {
        return td;
    }
    let name = team[i].Name;
    td.append($("<a>")
        .attr("href", "/search/?name=" + encodeURIComponent(name))
        .append(name));
    return td;
}

function fillShortMatchPlayerEloRow(team, i) {
    let td = $("<td>");
    if (i >= team.length) {
        return td;
    }
    td.append(team[i].ELO);
    return td;
}

function fillShortMatchDetailsRows(details) {
    let n_len = 0;
    let s_len = 0;
    if (details.South.Players != null && details.North.Players != null) {
        n_len = details.North.Players.length;
        s_len = details.South.Players.length;
    }
    let tbl_thead = $("<tr>")
            .addClass("table-secondary");
    if (showELO()) {
        tbl_thead.append($("<td>").append("North Team [" + n_len + "]"))
            .append($("<td>").append("ELO"))
            .append($("<td>").append("South Team [" + s_len + "]"))
            .append($("<td>").append("ELO"));
    } else {
        tbl_thead.append($("<td>").append("North Team [" + n_len + "]"))
            .append($("<td>").append("South Team [" + s_len + "]"));
    }
    let tbl = $("<table>")
        .addClass("table")
        .addClass("table-sm")
        .addClass("table-hover")
        .addClass("table-striped")
        .append($("<thead>").append(tbl_thead));

    let body = $("<tbody>");
    let n = Math.max(n_len, s_len);
    for (let i=0; i<n; i++) {
        let tr = $("<tr>");
        tr.append(fillShortMatchPlayerNameRow(details.North.Players, i));
        if (showELO()) {
            tr.append(fillShortMatchPlayerEloRow(details.North.Players, i));
        }
        tr.append(fillShortMatchPlayerNameRow(details.South.Players, i));
        if (showELO()) {
            tr.append(fillShortMatchPlayerEloRow(details.South.Players, i));
        }
        body.append(tr);
    }
    tbl.append(body);
    
    return tbl;
}

function getNiceTimeFromat(since) {
    let seconds = "" + since.seconds();
    if (seconds < 10) {
        seconds = "0" + seconds;
    }
    let minutes = "" + (60*(since.hours())+since.minutes());
    if (minutes < 10) {
        minutes = "0" + minutes;
    }
    return (minutes + ":" + seconds);
}

function fillLongMatchDetailsRows(details) {
    let tbl = $("<table>")
        .addClass("table-sm")
        .addClass("table-hover")
        .addClass("table-striped")
        .append($("<thead>").append($("<tr>")
            .addClass("table-secondary")
            .append($("<td>").append("Time"))
            //.append($("<td>").append("Event"))
            .append($("<td>").append("Info"))
        ));
    
    let body = $("<tbody>");
    let ref = moment(details.Start);
    $.each(details.Timeline, function(index, data) {
        // Ignore some events for now.
        if (data.EventType < 4) {
            return;
        }
        let since = moment.duration(moment(data.Timestamp) - ref, 'ms');
        body.append($("<tr>")
            .append($("<td>").append(getNiceTimeFromat(since)))
            //.append($("<td>").append(data.EventType))
            .append($("<td>").append(data.Payload))
        );
    })
    tbl.append(body);
    return tbl;
}

function getMVPTable(details) {
    let tbl = $("<table>")
        .addClass("table")
        .addClass("table-hover")
        .addClass("table-striped")
        .append($("<thead>")
            .addClass("table-secondary").append($("<tr>")
            .append($("<td>").append(""))
            .append($("<td>").append("Defenders").attr("colSpan", "3"))
            .append($("<td>").append("Deaths").attr("colSpan", "3"))
            .append($("<td>").append("Builders").attr("colSpan", "3"))
        ).append($("<tr>")
            .append($("<td>").append("#"))
            .append($("<td>").append("Name"))
            .append($("<td>").append("Count"))
            .append($("<td>").append("Total"))
            .append($("<td>").append("Name"))
            .append($("<td>").append("Count"))
            .append($("<td>").append("Total"))
            .append($("<td>").append("Name"))
            .append($("<td>").append("Count"))
            .append($("<td>").append("Total"))
        ));
    
    let body = $("<tbody>");
    let n = Math.min(details.Defenders.length, details.Deaths.length, details.Builders.length)
    for (let i=0; i<n; i++) {
        let def = details.Defenders[i];
        let ded = details.Deaths[i];
        let bui = details.Builders[i];
        body.append($("<tr>")
            .append($("<td>").append(i+1))
            .append($("<td>").append($("<a>")
                .attr("href", "/search/?name=" + encodeURIComponent(def.Name))
                .append(def.Name)))
            .append($("<td>").append(def.Stat))
            .append($("<td>").append(def.Total))
            .append($("<td>").append($("<a>")
                .attr("href", "/search/?name=" + encodeURIComponent(ded.Name))
                .append(ded.Name)))
            .append($("<td>").append(ded.Stat))
            .append($("<td>").append(ded.Total))
            .append($("<td>").append($("<a>")
                .attr("href", "/search/?name=" + encodeURIComponent(bui.Name))
                .append(bui.Name)))
            .append($("<td>").append(bui.Stat))
            .append($("<td>").append(bui.Total))
        );
    }
    tbl.append(body);
    return tbl;
}

function getMatchDetails(event) {
    let id = event.data.ID;
    if (id == 0) {
        return;
    }
    let tr = $("<tr>")
        .attr("colSpan", "5")
        .hide();
    $(this).after(tr);
    tr.fadeIn('fast');
    let endpoint = "/api/match/short/";
    let isLong = $("#det-long").is(':checked');
    if (isLong) {
        endpoint = "/api/match/long/";
    }
    $.getJSON(endpoint + id)
        .done(function(data) {
            let tbl;
            if (isLong) {
                tbl = fillLongMatchDetailsRows(data);
            } else {
                tbl = fillShortMatchDetailsRows(data);
            }
            tr.append($("<td>")
                .append($("<small>").append(tbl))
                .attr("colSpan", "5")
                .attr("align", "center")
            );
        })
        .fail(function( jqxhr, textStatus, error ) {
            var err = textStatus + ", " + error;
            console.log( "Request Failed: " + err );
            tr.append($("<td>")
                .append("Fetching recent match details failed: " + err));
        });
    $(this).one("click", {ID: id}, function(event) {
        tr.fadeOut('fast', function() {
            tr.remove();
        });
        $(this).one("click", {ID: event.data.ID}, getMatchDetails);
    });
}

function addRecentMatchesEntry(tbl, match) {
    let ago = moment(match.Start);
    let winner = "South";
    if (match.NorthWon) {
      winner = "North";
    }
    let diffStr = diff2str[match.Difficulty];
    let row = $("<tr>")
      .append($("<td>").append($("<a>")
                .attr("href", "/match/?id=" + match.ID)
                .append(match.ID)))
      .append($("<td>").append(ago.fromNow()))
      .append($("<td>").append(moment.duration(match.Length/1000000).humanize()))
      .append($("<td>").append(winner))
      .append($("<td>").append(diffStr))
      .one("click", {ID: match.ID}, getMatchDetails);
    if (!(typeof match.IsWinner === 'undefined')) {
        if (match.IsWinner) {
            row.addClass("table-success")
        } else {
            row.addClass("table-danger")
        }
    }
    tbl.append(row);
}

function showELO() {
    let elo = urlParams.get("elo");
    if (elo == null) {
        return false;
    }
    return elo == "1";
}
