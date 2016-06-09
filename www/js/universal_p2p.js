'use strict';

function isTorrentFile (file) {
  return file.name.toLowerCase().endsWith('.torrent')
}

function onTorrent(files) {
  console.log('got files');

  if(!isTorrentFile(files[0]) || files.length>1){
    alert("you must only select 1 .torrent file")
    return;
  }
  
  //send form data
  
  $("#torrent_input").prop("files",files)
  $("#torrent_form").submit()


}

var isAdvancedUpload = function() {
  var div = document.createElement('div');
  return (('draggable' in div) || ('ondragstart' in div && 'ondrop' in div)) 
    && 'FormData' in window && 'FileReader' in window;}();

function init(){

  $("html").on("dragover", function(event) {
    event.preventDefault();  
    event.stopPropagation();
  });

  $("html").on("dragleave", function(event) {
    event.preventDefault();  
    event.stopPropagation();
  });

  $("html").on("drop", function(event) {
    event.preventDefault();  
    event.stopPropagation();
  });

  
  $("#hidden_iframe").load(function() {
    alert("complete");
   });

  $("#torrent_input").change(function(){
    $("#torrent_form").submit()
  });




  [ 'drag', 'dragstart', 'dragend', 'dragover', 'dragenter', 'dragleave', 'drop' ].forEach( function( event )
      {
        $("#torrent_form").bind( event, function( e )
            {
              // preventing the unwanted behaviours
              e.preventDefault();
              e.stopPropagation();
            });
      });
  [ 'dragover', 'dragenter' ].forEach( function( event )
      {
        $("#torrent_form").bind( event, function()
            {
              $("#torrent_form").addClass( 'is-dragover' );
            });
      });
  [ 'dragleave', 'dragend', 'drop' ].forEach( function( event )
      {
        $("#torrent_form").bind( event, function()
            {
              $("#torrent_form").removeClass( 'is-dragover' );
            });
      });

  $("#torrent_form").bind( 'drop', function( e )
      {
        onTorrent( e.originalEvent.dataTransfer.files);
      });


  if (isAdvancedUpload) {
    $("#torrent_form").addClass('has-advanced-upload');
  }

}

$(function(){
  init()
});
