# API Testing Guide - Meeting Slot Recommendation System

This guide demonstrates how to use the Meeting Slot Service API to solve the problem of finding optimal meeting times for geographically distributed teams.

## Problem Statement

**Challenge**: In a geographically distributed team, it is very hard to find common time to meet that works for everyone.

**Solution**: This API helps organizers:
1. Create events with multiple proposed time slots
2. Collect availability from all participants
3. Get smart recommendations for meeting times that work for everyone
4. Get fallback options when perfect alignment isn't possible

---

## Prerequisites

- API server running at `http://localhost:8080` (or your deployed ALB URL)
- API testing tool (Postman, Thunder Client, or similar)
- Basic understanding of REST APIs and JSON

## How to Use This Guide

Each API call shows:
1. **HTTP Method + Endpoint** (e.g., `POST /api/v1/users`)
2. **Request JSON** - Copy this payload and send it to the endpoint

Simply copy the JSON request body and use it in your API testing tool.

---

## Step-by-Step API Test Scenario

### Scenario Description

**Event**: "Brainstorming Meeting" for Q1 Planning  
**Organizer**: Sarah Johnson (Product Manager, Pacific Time)  
**Participants**: 
- Sarah Johnson (San Francisco, PST/UTC-8)
- Raj Kumar (Bangalore, IST/UTC+5:30)
- Emma Schmidt (Berlin, CET/UTC+1)
- Carlos Rivera (New York, EST/UTC-5)

**Proposed Slots by Organizer** (all stored as UTC):

| Slot | UTC Time | PST | IST | CET | EST |
|------|----------|-----|-----|-----|-----|
| **1** | Jan 15, 22:00 - Jan 16, 01:00 | Jan 15, 2:00 PM - 5:00 PM | Jan 16, 3:30 AM - 6:30 AM | Jan 15, 11:00 PM - 2:00 AM | Jan 15, 5:00 PM - 8:00 PM |
| **2** | Jan 16, 16:00 - 19:00 | Jan 16, 8:00 AM - 11:00 AM | Jan 16, 9:30 PM - 12:30 AM | Jan 16, 5:00 PM - 8:00 PM | Jan 16, 11:00 AM - 2:00 PM |
| **3** | Jan 18, 02:00 - 05:00 | Jan 17, 6:00 PM - 9:00 PM | Jan 18, 7:30 AM - 10:30 AM | Jan 18, 3:00 AM - 6:00 AM | Jan 17, 9:00 PM - 12:00 AM |

**Meeting Duration Required**: 90 minutes (1.5 hours)

**Target**: Find a 90-minute slot where all 4 participants can meet. Slot 2 (16:00-17:30 UTC) is ideal for global team collaboration.

---

## API Testing Steps

### Step 1: Check API Health

Verify the API is running.

**Endpoint:** `GET /health`

**No request body required**

**Expected Response:**
```
OK
```

---

### Step 2: Create the Organizer User

Create Sarah Johnson (organizer).

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "name": "Sarah Johnson",
  "email": "sarah.johnson@company.com"
}
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": "usr_xxxxxxxxxxxx",
    "name": "Sarah Johnson",
    "email": "sarah.johnson@company.com",
    "created_at": "2026-01-14T10:00:00Z",
    "updated_at": "2026-01-14T10:00:00Z"
  }
}
```

**💡 Save the user ID** from `data.id` (use this in next steps)

---

### Step 3: Create Additional Participants

**3.1 Create Raj Kumar:**

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "name": "Raj Kumar",
  "email": "raj.kumar@company.com"
}
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": "usr_xxxxxxxxxxxx",
    "name": "Raj Kumar",
    "email": "raj.kumar@company.com",
    "created_at": "2026-01-14T10:01:00Z",
    "updated_at": "2026-01-14T10:01:00Z"
  }
}
```

**3.2 Create Emma Schmidt:**

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "name": "Emma Schmidt",
  "email": "emma.schmidt@company.com"
}
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": "usr_xxxxxxxxxxxx",
    "name": "Emma Schmidt",
    "email": "emma.schmidt@company.com",
    "created_at": "2026-01-14T10:02:00Z",
    "updated_at": "2026-01-14T10:02:00Z"
  }
}
```

**3.3 Create Carlos Rivera:**

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "name": "Carlos Rivera",
  "email": "carlos.rivera@company.com"
}
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": "usr_xxxxxxxxxxxx",
    "name": "Carlos Rivera",
    "email": "carlos.rivera@company.com",
    "created_at": "2026-01-14T10:03:00Z",
    "updated_at": "2026-01-14T10:03:00Z"
  }
}
```

**💡 Save all user IDs** from each response's `data.id` field. Example:
- Sarah (Organizer): `usr_4b1ef2a9465b`
- Raj: `usr_2ff3020a72ca`
- Emma: `usr_0fdb22368f5a`
- Carlos: `usr_2f57cd14efb7`

---

### Step 4: Create the Event with Proposed Slots

Sarah creates the "Brainstorming Meeting" event with 3 proposed time slots.

**Endpoint:** `POST /api/v1/events`

**Request:**
```json
{
  "title": "Q1 Brainstorming Meeting",
  "description": "Strategic planning session for Q1 initiatives. We need to align on priorities and resource allocation.",
  "organizer_id": "<SARAH_USER_ID>",
  "duration_minutes": 90,
  "proposed_slots": [
    {
      "start_time": "2026-01-15T22:00:00Z",
      "end_time": "2026-01-16T01:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "start_time": "2026-01-16T16:00:00Z",
      "end_time": "2026-01-16T19:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "start_time": "2026-01-18T02:00:00Z",
      "end_time": "2026-01-18T05:00:00Z",
      "timezone": "America/Los_Angeles"
    }
  ]
}
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": "evt_xxxxxxxxxxxx",
    "title": "Q1 Brainstorming Meeting",
    "description": "Strategic planning session for Q1 initiatives. We need to align on priorities and resource allocation.",
    "organizer_id": "<SARAH_USER_ID>",
    "duration_minutes": 90,
    "status": "pending",
    "proposed_slots": [
      {
        "start_time": "2026-01-15T22:00:00Z",
        "end_time": "2026-01-16T01:00:00Z",
        "timezone": "America/Los_Angeles"
      },
      {
        "start_time": "2026-01-16T16:00:00Z",
        "end_time": "2026-01-16T19:00:00Z",
        "timezone": "America/Los_Angeles"
      },
      {
        "start_time": "2026-01-18T02:00:00Z",
        "end_time": "2026-01-18T05:00:00Z",
        "timezone": "America/Los_Angeles"
      }
    ],
    "participants": [],
    "created_at": "2026-01-14T10:05:00Z",
    "updated_at": "2026-01-14T10:05:00Z"
  }
}
```

**💡 Save the event ID** from `data.id`

---

### Step 5: Add Participants to the Event

Add all participants to the event in a single API call.

**Endpoint:** `POST /api/v1/events/<EVENT_ID>/participants`

**Request:**
```json
{
  "user_ids": [
    "<SARAH_USER_ID>",
    "<RAJ_USER_ID>",
    "<EMMA_USER_ID>",
    "<CARLOS_USER_ID>"
  ]
}
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "message": "Participants processed",
    "added_count": 4,
    "failed_count": 0,
    "added_user_ids": [
      "<SARAH_USER_ID>",
      "<RAJ_USER_ID>",
      "<EMMA_USER_ID>",
      "<CARLOS_USER_ID>"
    ],
    "failed": null
  }
}
```

---

### Step 6: Participants Submit Their Availability

Now each participant submits their available time windows. **Important:** To get a 100% match at 16:00-17:30 UTC (Jan 16), all participants must submit availability that covers this window.

**6.1 Sarah Johnson's Availability (Organizer - PST):**

Sarah is available for all three proposed slots.
- Slot 1: Jan 15, 2:00 PM - 5:00 PM PST (22:00-01:00 UTC)
- Slot 2: Jan 16, 8:00 AM - 11:00 AM PST (16:00-19:00 UTC) ← Target slot
- Slot 3: Jan 17, 6:00 PM - 8:30 PM PST (02:00-04:30 UTC next day)

**Endpoint:** `POST /api/v1/events/<EVENT_ID>/participants/<SARAH_USER_ID>/availability`

**Request:**
```json
{
  "available_slots": [
    {
      "start_time": "2026-01-15T22:00:00Z",
      "end_time": "2026-01-16T01:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "start_time": "2026-01-16T16:00:00Z",
      "end_time": "2026-01-16T19:00:00Z",
      "timezone": "America/Los_Angeles"
    },
    {
      "start_time": "2026-01-18T02:00:00Z",
      "end_time": "2026-01-18T04:30:00Z",
      "timezone": "America/Los_Angeles"
    }
  ]
}
```

**6.2 Raj Kumar's Availability (Bangalore - IST):**

Raj extends his availability to late evening to accommodate the global team.
- 16:00 UTC = 9:30 PM IST, 17:30 UTC = 11:00 PM IST
- Available: 9:00 PM - midnight IST (15:30-18:30 UTC)

**Endpoint:** `POST /api/v1/events/<EVENT_ID>/participants/<RAJ_USER_ID>/availability`

**Request:**
```json
{
  "available_slots": [
    {
      "start_time": "2026-01-16T15:30:00Z",
      "end_time": "2026-01-16T18:30:00Z",
      "timezone": "Asia/Kolkata"
    }
  ]
}
```

**6.3 Emma Schmidt's Availability (Berlin - CET):**

Emma is available from 4:00 PM - 7:00 PM CET (end of work day + overtime).
- 16:00 UTC = 5:00 PM CET, 17:30 UTC = 6:30 PM CET
- Available: 15:00-18:00 UTC (4:00 PM - 7:00 PM CET)

**Endpoint:** `POST /api/v1/events/<EVENT_ID>/participants/<EMMA_USER_ID>/availability`

**Request:**
```json
{
  "available_slots": [
    {
      "start_time": "2026-01-16T15:00:00Z",
      "end_time": "2026-01-16T18:00:00Z",
      "timezone": "Europe/Berlin"
    }
  ]
}
```

**6.4 Carlos Rivera's Availability (New York - EST):**

Carlos is flexible and available during business hours EST.
- 16:00 UTC = 11:00 AM EST, 17:30 UTC = 12:30 PM EST
- Available: 10:00 AM - 3:00 PM EST (15:00-20:00 UTC)

**Endpoint:** `POST /api/v1/events/<EVENT_ID>/participants/<CARLOS_USER_ID>/availability`

**Request:**
```json
{
  "available_slots": [
    {
      "start_time": "2026-01-16T15:00:00Z",
      "end_time": "2026-01-16T20:00:00Z",
      "timezone": "America/New_York"
    }
  ]
}
```

**Expected Response (for each availability submission):**
```json
{
  "success": true,
  "data": {
    "message": "Availability submitted successfully"
  }
}
```

---

### Step 7: Get Recommendations

Now request the system to analyze all availabilities and recommend optimal meeting times.

**Endpoint:** `GET /api/v1/events/<EVENT_ID>/recommendations`

**No request body required**

**Expected Response:**

The system returns the best recommendation based on maximum availability.

```json
{
  "success": true,
  "data": {
    "event_id": "<EVENT_ID>",
    "duration_minutes": 90,
    "total_participants": 4,
    "best_recommendation": {
      "slot": {
        "start_time": "2026-01-16T08:00:00-08:00",
        "end_time": "2026-01-16T09:30:00-08:00",
        "timezone": "America/Los_Angeles"
      },
      "available_participants": 4,
      "availability_rate": 1,
      "available_users": [
        "<SARAH_USER_ID>",
        "<RAJ_USER_ID>",
        "<EMMA_USER_ID>",
        "<CARLOS_USER_ID>"
      ],
      "unavailable_users": []
    },
    "message": "Perfect match! All 4 participants are available for this time slot"
  }
}
```

**Note:** The time is displayed in the organizer's timezone (America/Los_Angeles). The UTC equivalent is 16:00-17:30 UTC.

---

## Understanding the Results

### ✅ Best Recommendation
**Time**: January 16, 2026, 8:00 AM - 9:30 AM PST (90 minutes)

| Participant | Local Time | Status |
|-------------|------------|--------|
| **Sarah** | 8:00 AM PST | ✅ Working hours |
| **Raj** | 9:30 PM IST | ✅ Late evening |
| **Emma** | 5:00 PM CET | ✅ End of work day |
| **Carlos** | 11:00 AM EST | ✅ Mid-morning |

**Result**: **100% availability** - Works for everyone! 🎉

---

## Additional API Operations

### Get Event Details

**Endpoint:** `GET /api/v1/events/<EVENT_ID>`

**No request body required**

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": "evt_xxxxxxxxxxxx",
    "title": "Q1 Brainstorming Meeting",
    "description": "Strategic planning session for Q1 initiatives.",
    "organizer_id": "<SARAH_USER_ID>",
    "duration_minutes": 90,
    "status": "pending",
    "proposed_slots": [
      {
        "start_time": "2026-01-15T22:00:00Z",
        "end_time": "2026-01-16T01:00:00Z",
        "timezone": "America/Los_Angeles"
      }
    ],
    "participants": [
      {
        "status": "responded",
        "user": {
          "id": "<SARAH_USER_ID>",
          "name": "Sarah Johnson",
          "email": "sarah.johnson@company.com",
          "created_at": "2026-01-14T10:00:00Z",
          "updated_at": "2026-01-14T10:00:00Z"
        }
      }
    ],
    "created_at": "2026-01-14T10:05:00Z",
    "updated_at": "2026-01-14T10:05:00Z"
  }
}
```

### List All Participants

**Endpoint:** `GET /api/v1/events/<EVENT_ID>/participants`

**No request body required**

**Expected Response:**
```json
{
  "success": true,
  "data": [
    {
      "status": "responded",
      "user": {
        "id": "<SARAH_USER_ID>",
        "name": "Sarah Johnson",
        "email": "sarah.johnson@company.com",
        "created_at": "2026-01-14T10:00:00Z",
        "updated_at": "2026-01-14T10:00:00Z"
      }
    },
    {
      "status": "responded",
      "user": {
        "id": "<RAJ_USER_ID>",
        "name": "Raj Kumar",
        "email": "raj.kumar@company.com",
        "created_at": "2026-01-14T10:01:00Z",
        "updated_at": "2026-01-14T10:01:00Z"
      }
    }
  ]
}
```

### Get Participant's Availability

**Endpoint:** `GET /api/v1/events/<EVENT_ID>/participants/<USER_ID>/availability`

**No request body required**

### Update Availability

**Endpoint:** `PUT /api/v1/events/<EVENT_ID>/participants/<USER_ID>/availability`

**Request:**
```json
{
  "available_slots": [
    {
      "start_time": "2026-01-16T14:00:00Z",
      "end_time": "2026-01-16T20:00:00Z",
      "timezone": "America/Los_Angeles"
    }
  ]
}
```

### Remove Participant

**Endpoint:** `DELETE /api/v1/events/<EVENT_ID>/participants/<USER_ID>`

**No request body required**

### Delete Event

**Endpoint:** `DELETE /api/v1/events/<EVENT_ID>`

**No request body required**

---