#!/bin/bash

# Real-time deployment monitoring script

echo "👀 Watching Deployments (Press Ctrl+C to stop)"
echo "=============================================="
echo ""

while true; do
    clear
    echo "📊 Deployment Status - $(date +%H:%M:%S)"
    echo "=========================================="
    echo ""
    
    # Latest deployment
    echo "🔄 Latest Deployment:"
    sqlite3 deployments.db "
    SELECT 
        '  ID: ' || d.id,
        '  Status: ' || d.status,
        '  Hostname: ' || COALESCE(d.hostname, 'Not assigned yet'),
        '  Commit: ' || substr(d.commit_sha, 1, 7),
        '  Created: ' || datetime(d.created_at, 'localtime')
    FROM deployments d 
    JOIN projects p ON d.project_id = p.id 
    WHERE p.user_id = 2 
    ORDER BY d.id DESC 
    LIMIT 1;
    " 2>/dev/null | sed 's/^/  /'
    
    echo ""
    echo "📦 Recent Deployments (Last 5):"
    sqlite3 deployments.db "
    SELECT 
        '  [' || d.id || '] ' || 
        d.status || ' | ' || 
        COALESCE(d.hostname, 'no-hostname') || ' | ' ||
        substr(d.commit_sha, 1, 7) || ' | ' ||
        datetime(d.created_at, 'localtime')
    FROM deployments d 
    JOIN projects p ON d.project_id = p.id 
    WHERE p.user_id = 2 
    ORDER BY d.id DESC 
    LIMIT 5;
    " 2>/dev/null | sed 's/^/  /'
    
    echo ""
    echo "🌐 Hostname Status:"
    HOSTNAMES=$(sqlite3 deployments.db "
    SELECT DISTINCT hostname 
    FROM deployments d 
    JOIN projects p ON d.project_id = p.id 
    WHERE p.user_id = 2 AND hostname != '' 
    ORDER BY d.id DESC 
    LIMIT 3;
    " 2>/dev/null)
    
    if [ -n "$HOSTNAMES" ]; then
        echo "$HOSTNAMES" | while read hostname; do
            echo "  ✅ $hostname"
        done
    else
        echo "  ⚠️  No hostnames assigned yet"
    fi
    
    echo ""
    echo "=========================================="
    echo "💡 Push code to GitHub to trigger deployment"
    echo "   Watching for changes... (Refresh every 2s)"
    
    sleep 2
done
