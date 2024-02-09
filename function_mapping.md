### Mapping: 

connect_db():
    connect to database SQLite


init_db():
    creates database tables

query_db(query, args=(), one=False):
    quesies database entries and returns the list of dictonaries


get_user_id(username):
    """Convenience method to look up the id for a username."""

format_datetime(timestamp):
    """Format a timestamp for display."""

gravatar_url(email, size=80):
    """Return the gravatar image for the given email address."""

@app.before_request
 before_request():
    Make sure we are connected to the database each request and look
    up the current user so that we know he's there.
    

@app.after_request
def after_request(response):
    """Closes the database again at the end of the request."""

@app.route('/')
def timeline():
    """Shows a users timeline or if no user is logged in it will
    redirect to the public timeline.  This timeline shows the user's
    messages as well as all the messages of followed users.
    """

@app.route('/public')
def public_timeline():
    """Displays the latest messages of all users."""

@app.route('/<username>')
    """Display's a users tweets."""
