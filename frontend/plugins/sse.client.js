import { defineNuxtPlugin } from '#app'

export default defineNuxtPlugin((nuxtApp) => {
  const eventSource = new EventSource(`http://localhost:8077/api/events`);
  const subscribers = new Set();

  eventSource.onmessage = function(event) {
    try {
      const logEvent = JSON.parse(event.data);
      subscribers.forEach(callback => callback(logEvent));
    } catch (error) {
      console.error('Error parsing log event:', error);
    }
  };

  eventSource.onerror = function(error) {
    console.error('SSE error:', error);
    eventSource.close();
  };

  const subscribeToLogs = (callback) => {
    subscribers.add(callback);
    return () => subscribers.delete(callback);
  };

  return {
    provide: {
      eventSource,
      subscribeToLogs
    },
  };
});
