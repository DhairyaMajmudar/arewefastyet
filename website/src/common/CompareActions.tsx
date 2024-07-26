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

import { Button } from "@/components/ui/button";
import { VitessRefs } from "@/types";
import React from "react";
import VitessRefsCommand from "./VitessRefsCommand";

export type CompareActionsProps = {
  oldGitRef: string;
  setOldGitRef: React.Dispatch<React.SetStateAction<string>>;
  newGitRef: string;
  setNewGitRef: React.Dispatch<React.SetStateAction<string>>;
  compareClicked: () => void;
  vitessRefs: VitessRefs | null;
};

export default function CompareAction(props: CompareActionsProps) {
  const {
    newGitRef,
    oldGitRef,
    setNewGitRef,
    setOldGitRef,
    compareClicked,
    vitessRefs,
  } = props;
  const isButtonDisabled = !oldGitRef || !newGitRef;

  return (
    <div className="flex flex-col md:flex-row gap-4">
      <div className="flex flex-col">
        <label className="text-primary mb-2">Old</label>
        <VitessRefsCommand
          inputLabel={"Search commit or releases..."}
          gitRef={oldGitRef}
          setGitRef={setOldGitRef}
          vitessRefs={vitessRefs}
          keyboardShortcut="o"
        />
      </div>
      <div className="flex flex-col">
        <label className="text-primary mb-2">New</label>
        <VitessRefsCommand
          inputLabel={"Search commit or releases..."}
          gitRef={newGitRef}
          setGitRef={setNewGitRef}
          vitessRefs={vitessRefs}
          keyboardShortcut="j"
        />
      </div>
      <div className="flex md:items-end items-center justify-center mt-4 md:mt-0">
        <Button
          onClick={compareClicked}
          disabled={isButtonDisabled}
          className="w-fit md:w-auto"
        >
          Compare
        </Button>
      </div>
    </div>
  );
}
