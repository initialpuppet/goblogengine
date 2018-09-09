var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

// Font Awesome
gulp.task('fontawesome', function () {
  return gulp.src('node_modules/font-awesome/fonts/*').pipe(gulp.dest('main/static/fonts')) &&
    gulp.src('node_modules/font-awesome/css/font-awesome.min.css').pipe(gulp.dest('main/static/css'));
})

// SimpleMDE
gulp.task('simplemde', function () {
  return gulp.src('node_modules/simplemde/dist/simplemde.min.js').pipe(gulp.dest('main/static/js')) &&
    gulp.src('node_modules/simplemde/dist/simplemde.min.css').pipe(gulp.dest('main/static/css'));
})

// Tagify
// Uses local vesion incorporating bugfixes for Firefox / Edge
// TODO: merge the Firefox / Edge bugfixes upstream and pull from the NPM version
gulp.task('tagify', function () {
  return gulp.src('external/tagify/src/tagify.js').pipe(gulp.dest('main/static/js')) &&
    gulp.src('external/tagify/dist/tagify.css').pipe(gulp.dest('main/static/css'));
})

// JavaScript libraries
// TODO: Select only Foundation modules in use
var jslibPaths = [
  'node_modules/foundation-sites/dist/js/foundation.min.js',
  'node_modules/foundation-sites/vendor/jquery/dist/jquery.min.js',
  'node_modules/blueimp-file-upload/js/jquery.fileupload.js',
  'node_modules/blueimp-file-upload/js/jquery.iframe-transport.js',
  'node_modules/blueimp-file-upload/js/vendor/jquery.ui.widget.js',
  'node_modules/blueimp-tmpl/js/tmpl.min.js'
];
gulp.task('jslib', function () {
  return gulp.src(jslibPaths)
    .pipe(gulp.dest('main/static/js/'));
});

// Concatenate and minify application JavaScript
gulp.task('js', function () {
  return gulp.src('main/assets/js/app.js')
    .pipe($.uglify())
    .pipe($.concat('app.min.js'))
    .pipe(gulp.dest('main/static/js/'));
});

// Build and concatenate SCSS files
var sassPaths = [
  'node_modules/foundation-sites/scss'
];
gulp.task('sass', function () {
  return gulp.src('main/assets/scss/app.scss')
    .pipe($.sass({
      includePaths: sassPaths,
      outputStyle: 'compressed'
    })
      .on('error', $.sass.logError))
    .pipe($.autoprefixer({
      browsers: ['last 2 versions', 'ie >= 9']
    }))
    .pipe(gulp.dest('main/static/css'));
});

// Images
// TODO: compression etc
gulp.task('images', function () {
  return gulp.src('main/assets/images/*').pipe(gulp.dest('main/static/images'));
})

// Default task
gulp.task('default', ['jslib', 'sass', 'js', 'images', 'simplemde', 'fontawesome', 'tagify'], function () {
    gulp.watch(['main/assets/scss/**/*.scss'], ['sass']);
    gulp.watch(['main/assets/js/**/*.js'], ['js']);
});

// Deployment task (no watching)
gulp.task('buildfordeploy', ['jslib', 'sass', 'js', 'images', 'simplemde', 'fontawesome', 'tagify']);