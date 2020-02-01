var diff2str = [
  "Peaceful",
  "Piece of Cake",
  "Easy",
  "Normal",
  "Hard",
  "Nightmare",
  "Insane"
  ];

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

function fillShortMatchDetailsRows(details) {
    let n_len = 0;
    let s_len = 0;
    if (details.South.Players != null && details.North.Players != null) {
        n_len = details.North.Players.length;
        s_len = details.South.Players.length;
    }
    let tbl = $("<table>")
        .addClass("table-sm")
        .append($("<thead>").append($("<tr>")
            .addClass("table-secondary")
            .append($("<td>").append("North Team [" + n_len + "]"))
            .append($("<td>").append("South Team [" + s_len + "]"))
        ));

    let n = Math.max(n_len, s_len);
    for (let i=0; i<n; i++) {
        let north = $("<td>");
        if (i < n_len) {
            let name = details.North.Players[i].Name;
            north.append($("<a>")
                .attr("href", "/site/search/?name=" + encodeURIComponent(name))
                .append(name));
        }
        let south = $("<td>");
        if (i < s_len) {
            let name = details.South.Players[i].Name;
            south.append($("<a>")
                .attr("href", "/site/search/?name=" + encodeURIComponent(name))
                .append(name));
        }
        tbl.append($("<tr>")
            .append(north)
            .append(south)
        );
    }
    
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
        .append($("<thead>").append($("<tr>")
            .addClass("table-secondary")
            .append($("<td>").append("Time"))
            //.append($("<td>").append("Event"))
            .append($("<td>").append("Info"))
        ));
    
    let ref = moment(details.Start);
    $.each(details.Timeline, function(index, data) {
        // Ignore some events for now.
        if (data.EventType < 4) {
            return;
        }
        let since = moment.duration(moment(data.Timestamp) - ref, 'ms');
        tbl.append($("<tr>")
            .append($("<td>").append(getNiceTimeFromat(since)))
            //.append($("<td>").append(data.EventType))
            .append($("<td>").append(data.Payload))
        );
    })
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
      .append($("<td>").append(match.ID))
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
