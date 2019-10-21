<html>
    <head>
    <title></title>
    </head>
    <body>
        {{range .Switches}}
            Name: {{.Name}}
            Description: {{.Description}}
        {{end}}
        <form action="/" method="post">
            Add new resource:<br/>
            Name:<input type="text" name="Name"><br/>
            Description:<input type="text" name="Description"><br>
            <input type="submit" value="Gotcha">
        </form>
    </body>
</html>
