// load_test_safe_active.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = 'http://localhost:8080';

export const options = {
  scenarios: {
    stage1_manage_users: {
      executor: 'constant-vus',
      vus: 20,
      duration: '60s',
      startTime: '5s',
      exec: 'manageUser',
    },
    
    stage2_create_prs: {
      executor: 'ramping-vus',
      startVUs: 10,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 100 },
        { duration: '30s', target: 0 },
      ],
      startTime: '70s',
      exec: 'createPR',
    },
  },
  
  thresholds: {
    http_req_duration: ['p(95)<800'],
    http_req_failed: ['rate<0.05'],
  },
};

export function setup() {
  console.log('\n=== SETUP: Creating Teams and Users ===\n');
  
  const allUsers = [];
  const safeUsers = []; 
  const teamsCreated = [];
  
  for (let i = 0; i < 20; i++) {
    const teamName = `LoadTest-Team-${i}`;
    const membersCount = Math.floor(Math.random() * 6) + 8;
    
    const members = [];
    for (let j = 0; j < membersCount; j++) {
      const userId = uuidv4();
      members.push({
        user_id: userId,
        username: `User_T${i}_M${j}`,
        is_active: true,
      });
      allUsers.push(userId);
      
      if (allUsers.length <= 100) {
        safeUsers.push(userId);
      }
    }
    
    const res = http.post(`${BASE_URL}/team/add`, JSON.stringify({
      team_name: teamName,
      members: members,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (res.status === 201) {
      teamsCreated.push(teamName);
      console.log(`✓ Team ${i + 1}/20: ${teamName} (${members.length} users, Total: ${allUsers.length})`);
    }
    
    sleep(0.3);
  }
  
  console.log(`\n=== SETUP COMPLETED ===`);
  console.log(`Total users: ${allUsers.length}`);
  console.log(`Protected users (always active): ${safeUsers.length}`);
  console.log(`Users for activation/deactivation tests: ${allUsers.length - safeUsers.length}\n`);
  
  return {
    allUserIds: allUsers,
    safeUserIds: safeUsers, 
    testUserIds: allUsers.slice(100), 
    teams: teamsCreated,
  };
}

export function manageUser(data) {
  if (!data || !data.testUserIds || data.testUserIds.length === 0) {
    const userId = data.allUserIds[Math.floor(Math.random() * data.allUserIds.length)];
    activateDeactivateUser(userId);
    return;
  }
  
  const userId = data.testUserIds[Math.floor(Math.random() * data.testUserIds.length)];
  activateDeactivateUser(userId);
}

function activateDeactivateUser(userId) {
  const res = http.post(`${BASE_URL}/users/setIsActive`, JSON.stringify({
    user_id: userId,
    is_active: Math.random() > 0.5,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(res, { 
    '[Users] Updated': (r) => r.status === 200 || r.status === 404,
  });
  
  sleep(Math.random() * 0.5 + 0.3);
}
export function createPR(data) {
  if (!data || !data.safeUserIds || data.safeUserIds.length === 0) {
    console.error('[PRs] No safe users available!');
    sleep(2);
    return;
  }
  
  const userId = data.safeUserIds[Math.floor(Math.random() * data.safeUserIds.length)];
  
  const res = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify({
    pull_request_id: uuidv4(),
    pull_request_name: `LoadTest-PR-${Date.now()}-${Math.random().toString(36).substring(7)}`,
    author_id: userId,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(res, { 
    '[PRs] Created': (r) => r.status === 201,
    '[PRs] Fast': (r) => r.timings.duration < 500,
  });
  
  sleep(0.3);
}

export function teardown(data) {
  console.log(`\n=== TEST COMPLETED ===`);
  console.log(`Total users: ${data.allUserIds.length}`);
  console.log(`Safe users used for PRs: ${data.safeUserIds.length}`);
  console.log(`Total teams: ${data.teams.length}\n`);
}
