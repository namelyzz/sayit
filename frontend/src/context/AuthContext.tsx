"use client";

import { createContext, useContext, useState, useEffect, ReactNode } from "react";
import { apiClient } from "@/lib/api";

interface User {
  user_id: string;
  user_name: string;
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (username: string, password: string, rememberMe?: boolean) => Promise<void>;
  signup: (username: string, password: string, rePassword: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext< AuthContextType | undefined>(undefined);

function getStorage(remember: boolean) {
  return remember ? localStorage : sessionStorage;
}

function clearAllStorage() {
  localStorage.removeItem("token");
  localStorage.removeItem("user_id");
  localStorage.removeItem("user_name");
  sessionStorage.removeItem("token");
  sessionStorage.removeItem("user_id");
  sessionStorage.removeItem("user_name");
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check both localStorage and sessionStorage for existing session
    const token = localStorage.getItem("token") || sessionStorage.getItem("token");
    const userId = localStorage.getItem("user_id") || sessionStorage.getItem("user_id");
    const userName = localStorage.getItem("user_name") || sessionStorage.getItem("user_name");

    if (token && userId && userName) {
      apiClient.setToken(token);
      setUser({ user_id: userId, user_name: userName });
    }
    setLoading(false);
  }, []);

  const login = async (username: string, password: string, rememberMe: boolean = false) => {
    const response = await apiClient.login(username, password);
    const { user_id, user_name, token } = response.data;
    
    const storage = getStorage(rememberMe);
    storage.setItem("token", token);
    storage.setItem("user_id", user_id);
    storage.setItem("user_name", user_name);
    
    setUser({ user_id, user_name });
  };

  const signup = async (username: string, password: string, rePassword: string) => {
    await apiClient.signup(username, password, rePassword);
  };

  const logout = () => {
    apiClient.clearToken();
    clearAllStorage();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, signup, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
