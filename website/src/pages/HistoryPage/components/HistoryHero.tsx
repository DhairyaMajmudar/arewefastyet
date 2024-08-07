/*
Copyright 2024 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import Hero, { HeroProps } from "@/common/Hero";

const heroProps: HeroProps = {
  title: "History",
  description: (
    <>
      Easily navigate the different commits that were benchmarked by
      arewefastyet, allowing for faster exploration and bisection of performance
      regressions. Each row is a unique pair of commit SHA and benchmark source.{" "}
    </>
  ),
};

export default function HistoryHero() {
  return (
    <>
      <Hero title={heroProps.title} description={heroProps.description} />
    </>
  );
}
