$(function () {

   var conn;
   var msg = $("#msg");
   var log = $("#log");

   function appLog(msg) {
      var d = log[0];
      var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
      msg.append(log);
      if (doScroll) {
         d.scrollTop = d.scrollHeight - d.clientHeight;
      }
   }

   $("#form").submit(function () {
      if (!conn) {
         return false;
      }
      if (!msg.val()) {
         return false;
      }
      conn.send(msg.val());
      msg.val("");
      return false;
   });

   if (window["WebSocket"]) {
      conn = new WebSocket("ws://localhost:8080/ws");
      conn.onclose = function (evt) {
         appLog($("<div><b>Connection Closed.<\/b><\/div>"));

      };
      conn.onmessage = function (evt) {
         appLog($("<div/>").text(evt.data));
      }

   } else {
      appLog($("<div><b>Your browser does not support WebSockets.<\/b><\/div>"));
   }
});