#!/bin/bash
# Test script to check if API is working
# First, get a token by logging in, then use it here

echo "Testing /api/projects endpoint..."
echo "Make sure you're logged in and have a token in localStorage"
echo ""
echo "To test manually:"
echo "1. Open browser console on dashboard"
echo "2. Run: fetch('/api/projects', { headers: { 'Authorization': 'Bearer ' + localStorage.getItem('token') } }).then(r => r.json()).then(console.log)"
echo ""
echo "Or use curl (replace TOKEN with your JWT token):"
echo "curl -H 'Authorization: Bearer TOKEN' http://localhost:8080/api/projects"
