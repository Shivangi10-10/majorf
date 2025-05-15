# 🚀 WorkWise – Bridging Talent and Opportunity Through Smart Networking

> A Discord-based intelligent referral bot leveraging graph algorithms and automation to connect job seekers with credible referrers—simplifying modern recruitment.

---

## 📌 Overview

**WorkWise** is a smart networking bot built for Discord, aimed at democratizing access to job referrals. It empowers job seekers and employee referrers by automating the discovery, ranking, and connection process using the HITS algorithm (Hyperlink-Induced Topic Search). 

Instead of relying on cold outreach via platforms like LinkedIn, users interact with a bot to get ranked matches for potential referrers, complete with credibility scoring, structured communication, and feedback mechanisms—all inside a familiar chat environment.

---

## 👥 Team Members

- **Shivangi Suyash** – 9921103053  
- **Aditi Singh** – 9921103066  
- **Palak Agarwal** – 9921103093  
**Supervisor**: Ms. Akanksha Mehndiratta  
**Institution**: Jaypee Institute of Information Technology, Noida

---

## 🎯 Key Features

- 🧠 **Graph-based Recommendation** using the HITS algorithm  
- 💬 **Discord Bot Integration** via `discordgo`  
- 🗃️ **MongoDB-backed Data Model** for users, ratings, and connections  
- 📊 **Role-based Ranking** for accurate referrer matches  
- 🤖 **AI-powered Referral Message Generator**  
- 🧾 **Feedback & Ratings System** for trust and accountability  
- 🔄 **Real-time Networking** inside Discord servers

---

## 🛠️ Tech Stack

| Tool/Framework     | Purpose                             |
|--------------------|-------------------------------------|
| Golang             | Backend Bot Development             |
| MongoDB            | NoSQL Database for Persistent Storage |
| Discord API        | Communication Platform              |
| HITS Algorithm     | Graph-Based Referrer Ranking        |
| GoDotEnv           | Secure Environment Configuration    |

---

## 🧱 Architecture

- **Bot Layer**: Handles user commands and responses
- **Database Layer**: Manages users, connections, and referral ratings
- **Graph Layer**: Constructs and analyzes user networks
- **UI Layer** (optional): Visualization using vis.js (Graph rendering)

---

## 🧪 Testing Summary

| Test Type         | Status     |
|-------------------|------------|
| Unit Tests        | ✅ Passed   |
| Integration Tests | ✅ Passed   |
| Performance Tests | ✅ Passed (1000+ connections) |
| Security Checks   | ✅ Environment variable protection |
| Manual UI Tests   | ✅ Conducted for graph rendering |

---

## 📊 Referral Matching Logic

The bot ranks users based on:
1. **Authority Score** (Referral credibility)
2. **Hub Score** (Network connectedness)
3. **Role Weighting** (e.g., Manager > SDE1)

Implemented using the **HITS algorithm**, it ensures relevance, speed, and fairness.

---

## 🧭 Commands Guide

```bash
!register <role> <company>      # Register your profile
!connect <@user>                # Connect with another user
!find_referrer <company>       # Get a ranked referrer list
!rate_referrer <@user> <1-5>   # Rate your referral experience
!suggest                       # Submit a platform improvement idea
!help                          # Show all commands
