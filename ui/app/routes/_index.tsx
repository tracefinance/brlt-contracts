import type { MetaFunction } from "@remix-run/node";
import { redirect } from "@remix-run/node";

export function loader() {
  return redirect("/wallets");
}

export default function Index() {
  return null;
}
